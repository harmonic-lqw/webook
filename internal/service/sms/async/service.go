package async

import (
	"context"
	"math/rand"
	"sync/atomic"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/internal/service/sms"
	"webook/pkg/logger"
	"webook/pkg/logger/old"
)

const (
	ErrPercent = 0.1
)

type Service struct {
	svc sms.Service

	// 转异步，存储发短信请求的 repository
	repo repository.AsyncSmsRepository
	l    old.LoggerV1

	// 标志异步发送是否开启
	signAsync bool
	// 基于错误率判断是否转异步
	errCnt   int32
	reqCnt   int32
	initTime time.Time
	// 异步发送时，保持同步发送的流量
	keep int32
}

func NewService(svc sms.Service, repo repository.AsyncSmsRepository, l old.LoggerV1) *Service {
	res := &Service{
		svc:       svc,
		repo:      repo,
		l:         l,
		signAsync: false,
		initTime:  time.Now(),
		keep:      10,
	}
	go func() {
		res.StartAsyncCycle()
	}()
	return res
}

// StartAsyncCycle 循环地异步发送消息
// 使用了最简单的抢占式调度
func (s *Service) StartAsyncCycle() {
	for {
		s.AsyncSend()
		break
	}
}

func (s *Service) AsyncSend() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//defer cancel()
	// 尝试从数据库中拿到一个需要异步发送的消息
	// 这里采用抢占式，因为需要注意当部署很多实例的时候，防止它们都拿到了发送请求导致多次发送
	as, err := s.repo.PreemptWaitingSMS(ctx)
	cancel()
	switch err {
	// 成功从数据库中拿到一个等待发送的短信请求，执行发送
	case nil:
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		// 注意：这里不是 s.Send(...)
		// 这里的逻辑是 当成功从数据库拿到一个待发送的短信后，直接发送即可，不需要再判断是否需要异步了
		err = s.svc.Send(ctx, as.TplId, as.Args, as.Numbers...)
		if err != nil {
			s.l.Error("拿到了数据库中的异步消息，但是执行发送失败",
				logger.Error(err),
				logger.Int64("id", as.Id))
		}

		res := err == nil
		err = s.repo.ReportScheduleResult(ctx, as.Id, res)

		if err != nil {
			s.l.Error("执行发送成功，但是标记失败",
				logger.Error(err),
				logger.Bool("res", res),
				logger.Int64("id", as.Id))
		}
	// 数据库没有可以发送的短信请求
	case repository.ErrWaitingSMSNotFound:
		// 处理比较灵活
		// 这里设置为睡眠 5s
		time.Sleep(time.Second * 5)
	default:
		// 此时大概率是数据库出了问题，
		// 但是为了尽量运行，还是要继续
		// 可以稍微睡眠，也可以不睡眠，睡眠的话可以帮你规避掉短时间的网络抖动问题
		s.l.Error("抢占异步发送短信任务失败，系统错误",
			logger.Error(err))
		time.Sleep(time.Second)
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	if s.needAsyncV1() {
		err := s.repo.Add(ctx, domain.AsyncSms{
			TplId:   tplId,
			Args:    args,
			Numbers: numbers,
			// 初始化最大重试次数, retry_max 3 时，retry_cnt 0/1/2，达到最大重试次数就标记状态为失败
			RetryMax: 3,
		})
		return err
	} else {
		err := s.svc.Send(ctx, tplId, args, numbers...)
		if err != nil {
			atomic.AddInt32(&s.errCnt, 1)
		}
		atomic.AddInt32(&s.reqCnt, 1)
		return err
	}
}

func (s *Service) needAsync() bool {
	// 基于错误率来判断何时进行异步发送
	// 一段时间内，收到 err 的请求比率大于一个阈值，就转异步
	nowU := time.Now().UnixMilli()
	errCalculateDuration := nowU - s.initTime.UnixMilli()
	// 如果时间间隔小于 10 分钟就继续同步发送
	if errCalculateDuration > 10*time.Minute.Milliseconds() && float64(s.errCnt)/float64(s.reqCnt) > ErrPercent {
		return true
	}
	return false
}

// 考虑了何时退出异步
// 退出策略：
// 异步发送时，仍保留 10% 同步发送，2 分钟后计算错误率，如果正常，调整为 20，之后依次为 40/80/退出异步
func (s *Service) needAsyncV1() bool {
	if !s.signAsync {
		errCnt := atomic.LoadInt32(&s.errCnt)
		reqCnt := atomic.LoadInt32(&s.reqCnt)
		now := time.Now()
		// 至少计算 10 分钟的错误率，期间保持同步发送
		if s.initTime.Add(time.Minute*10).Before(now) && float64(errCnt)/float64(reqCnt) > ErrPercent {
			// 开启异步发送
			s.signAsync = true
			s.initTime = now
			atomic.StoreInt32(&s.errCnt, 0)
			atomic.StoreInt32(&s.reqCnt, 0)
		}
		return false
	} else {
		keep := atomic.LoadInt32(&s.keep)

		randNumber := rand.Intn(100) + 1
		if int32(randNumber) > keep { // 异步发送
			return true
		} else { // 继续同步发送
			errCnt := atomic.LoadInt32(&s.errCnt)
			reqCnt := atomic.LoadInt32(&s.reqCnt)
			now := time.Now()
			// 更新 keep 并判断是否退出异步发送
			if s.initTime.Add(time.Minute*2).Before(now) && float64(errCnt)/float64(reqCnt) <= ErrPercent {
				if atomic.CompareAndSwapInt32(&s.keep, keep, 2*keep) {
					atomic.StoreInt32(&s.errCnt, 0)
					atomic.StoreInt32(&s.reqCnt, 0)
					s.initTime = now
				}
				if s.keep >= 100 {
					// 退出异步发送
					s.signAsync = false
				}
			}
			return false
		}
	}
}
