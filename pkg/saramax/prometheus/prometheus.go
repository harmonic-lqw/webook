package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type Decorator struct {
}

func (d *Decorator) DecoratorProduceTime(start time.Time) {
	labels := []string{"topic", "partition"}
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "harmonic",
		Subsystem: "webook",
		Name:      "kafka" + "_produce_time",
		Help:      "统计向 kafka 发送数据的时间",
		ConstLabels: map[string]string{
			"instance_id": "my_instance",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, labels)
	prometheus.MustRegister(vector)
}

func (d *Decorator) DecoratorConsumeTime(start time.Time) {
	labels := []string{"topic", "partition"}
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "harmonic",
		Subsystem: "webook",
		Name:      "kafka" + "_consume_time",
		Help:      "统计向 kafka 发送数据的时间",
		ConstLabels: map[string]string{
			"instance_id": "my_instance",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, labels)
	prometheus.MustRegister(vector)
}
