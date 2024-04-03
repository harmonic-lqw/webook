package wego

import (
	"github.com/robfig/cron/v3"
	"webook/pkg/ginx"
	"webook/pkg/grpcx"
	"webook/pkg/saramax"
)

type App struct {
	GRPCServer *grpcx.Server
	WebServer  *ginx.Server
	Consumers  []saramax.Consumer
	Cron       *cron.Cron
}
