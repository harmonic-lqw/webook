package wrr

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
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

		conns = append(conns, &weightConn{
			SubConn:       sc,
			weight:        int(weight),
			currentWeight: int(weight),
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
				maxConn.currentWeight -= maxConn.weight
			}
			if err == nil && maxConn.currentWeight <= 2*total {
				maxConn.currentWeight += maxConn.weight
			}
		},
	}, nil
}

type weightConn struct {
	balancer.SubConn
	weight        int
	currentWeight int
}
