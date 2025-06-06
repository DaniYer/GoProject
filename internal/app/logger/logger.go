package logger

import (
	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

func Init() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	Log = logger.Sugar()
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
