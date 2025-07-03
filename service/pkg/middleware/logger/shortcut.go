package logger

import (
	"context"
	"fmt"

	"github.com/scienceol/studio/service/pkg/common/constant"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap/zapcore"
)

var (
	path             string
	maxSize                        = 50
	maxBackups                     = 10
	maxAge                         = 7
	compress                       = false
	traceLogMinLevel zapcore.Level = zapcore.InfoLevel
)

func GetCurLoggerLevel() string {
	return atomicLevel.Level().String()
}

func SetLoggerLevel(lvl string) {
	atomicLevel.SetLevel(getLoggerLevel(lvl))
}

func MustInit(logPath string, serviceEnv ServiceEnv) {
	if err := InitLogger(Log{
		Path:             "",
		MaxSize:          maxSize,
		MaxBackups:       maxBackups,
		MaxAge:           maxAge,
		Compress:         compress,
		TraceLogMinLevel: traceLogMinLevel,
	}, serviceEnv); err != nil {
		panic(fmt.Sprintf("init logger err: %+v", err))
	}
}

func InitLogger(conf Log, serviceEnv ServiceEnv) error {
	var err error
	if serviceEnv.Env == "" {
		serviceEnv.Env = constant.EnvDev
	}
	if serviceEnv.Platform == "" {
		serviceEnv.Platform = constant.Platform
	}
	if serviceEnv.Service == "" {
		serviceEnv.Service = "unkonw"
	}
	var opts []otelzap.Option
	opts = append(opts, otelzap.WithMinLevel(conf.TraceLogMinLevel), otelzap.WithStackTrace(true))
	switch serviceEnv.Env {
	case constant.EnvDev: // 开发环境
		InitStdOutCtxLogger(serviceEnv.Platform, serviceEnv.Service, opts...)
	case constant.EnvProd, constant.EnvUat, constant.EnvTest: // 测试生产环境
		InitCtxLogger(conf, serviceEnv.Platform, serviceEnv.Service, opts...)
	}
	return err
}

func Debugf(ctx context.Context, format string, v ...interface{}) {
	CtxLogger(ctx).Debug(fmt.Sprintf(format, v...))
}

func Infof(ctx context.Context, format string, v ...interface{}) {
	CtxLogger(ctx).Info(fmt.Sprintf(format, v...))
}

func Warnf(ctx context.Context, format string, v ...interface{}) {
	CtxLogger(ctx).Warn(fmt.Sprintf(format, v...))
}

func Errorf(ctx context.Context, format string, v ...interface{}) {
	CtxLogger(ctx).Error(fmt.Sprintf(format, v...))
}

func Fatalf(ctx context.Context, format string, v ...interface{}) {
	CtxLogger(ctx).Fatal(fmt.Sprintf(format, v...))
}

func Close() error {
	_ = ctxLogger.Sync()
	lumberjackLoggerClose()
	return nil
}

func IsInitialized() bool { return ctxLogger != nil }
