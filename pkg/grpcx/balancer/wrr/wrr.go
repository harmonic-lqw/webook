package wrr

import (
	"fmt"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

const Name = "custom_weighted_round_robin"

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &PickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*weightConn, 0, len(info.ReadySCs))
	for sc, sci := range info.ReadySCs {
		md, _ := sci.Address.Metadata.(map[string]any)
		weighVal, _ := md["weight"]
		// 可以忽略 ok 是因为无论哪一步出问题最终 weight 通过 float64 断言都会为 0
		weight, _ := weighVal.(float64)

		// assignment week16
		addr := sci.Address.Addr

		conns = append(conns, &weightConn{
			SubConn:       sc,
			weight:        int(weight),
			currentWeight: int(weight),

			// assignment week16
			addr: addr,
		})
	}
	return &Picker{
		conns: conns,
	}
}

type Picker struct {
	conns []*weightConn
	lock  sync.Mutex
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var total int
	var maxConn *weightConn
	//for _, c := range p.conns {
	//	total += c.weight
	//}
	//for _, c := range p.conns {
	//	c.currentWeight += c.weight
	//}
	//for _, c := range p.conns {
	//	if maxConn == nil || maxConn.currentWeight < c.currentWeight {
	//		maxConn = c
	//	}
	//}

	// 前三步可以合并
	for _, c := range p.conns {
		total += c.weight
		c.currentWeight += c.weight
		if maxConn == nil || maxConn.currentWeight < c.currentWeight {
			maxConn = c
		}
	}

	maxConn.currentWeight -= total

	return balancer.PickResult{
		SubConn: maxConn.SubConn,
		// 回调函数，表示用选中的那个 SubConn 是否成功
		Done: func(info balancer.DoneInfo) {
			err := info.Err
			if err != nil && maxConn.currentWeight > 0 {
				st, _ := status.FromError(err)
				if st.Code() == codes.ResourceExhausted {
					maxConn.currentWeight = maxConn.weight
					maxConn.weight = 0
				} else {
					maxConn.currentWeight -= maxConn.weight
				}
			}
			if err == nil && maxConn.currentWeight <= 2*total {
				maxConn.weight = maxConn.currentWeight
				maxConn.currentWeight += maxConn.weight
			}
		},
	}, nil

	// assignment week16
	//return balancer.PickResult{
	//	SubConn: maxConn.SubConn,
	//	// 回调函数，表示用选中的那个 SubConn 是否成功
	//	Done: func(info balancer.DoneInfo) {
	//		err := info.Err
	//		st, _ := status.FromError(err)
	//
	//		if st.Code() == codes.ResourceExhausted {
	//			// 触发限流
	//			// 降低权重
	//			maxConn.currentWeight -= maxConn.weight
	//			return
	//		} else if st.Code() == codes.Unavailable {
	//			// 触发熔断
	//			// 将该节点移出可用节点列表
	//			p.lock.Lock()
	//			defer p.lock.Unlock()
	//			p.removeWeightConn(&p.conns, maxConn)
	//
	//			// 移出后的操作...
	//			go func() {
	//				ticker := time.NewTicker(time.Second * 10)
	//				defer ticker.Stop()
	//				for {
	//					select {
	//					case <-ticker.C:
	//						// 发送健康请求到服务端
	//						cc, err := grpc.Dial(maxConn.addr,
	//							grpc.WithTransportCredentials(insecure.NewCredentials()))
	//						client := grpc2.NewUserServiceClient(cc)
	//						_, err = client.GetByID(context.Background(), &grpc2.GetByIDRequest{Id: 123})
	//
	//						// 如果返回正常响应，挪回可用节点列表
	//						if err == nil {
	//							p.conns = append(p.conns, &weightConn{
	//								SubConn:       maxConn,
	//								weight:        maxConn.weight,
	//								currentWeight: 0,
	//								addr:          maxConn.addr,
	//							})
	//							return
	//						}
	//					}
	//
	//				}
	//			}()
	//			return
	//		}
	//		if err != nil && maxConn.currentWeight > 0 {
	//			maxConn.currentWeight -= maxConn.weight
	//		}
	//		if err == nil && maxConn.currentWeight <= 2*total {
	//			maxConn.currentWeight += maxConn.weight
	//		}
	//	},
	//}, nil
}

type weightConn struct {
	balancer.SubConn
	weight        int
	currentWeight int
	addr          string
}

func (p *Picker) removeWeightConn(conns *[]*weightConn, conn *weightConn) {
	for i, c := range *conns {
		if c == conn {
			// 找到要删除的元素，使用切片操作删除它
			*conns = append((*conns)[:i], (*conns)[i+1:]...)
			return
		}
	}
	// 如果元素不在切片中，可以选择打印一条消息或进行其他处理
	fmt.Println("weightConn not found in the slice")
}
