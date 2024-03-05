package ioc

import (
	"webook/pkg/gormx/connpool"
	"webook/pkg/logger"
)

func InitDoubleWritePool(src SrcDB, dst DstDB, l logger.LoggerV1) *connpool.DoubleWritePool {
	return connpool.NewDoubleWritePool(src, dst, l)
}
