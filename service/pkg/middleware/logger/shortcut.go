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

type ServiceEnv struct {
	Platform string
	Service  string
	Env      string
}

type LogConfig struct {
	Path       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	LogLevel   string
	ServiceEnv
}

func GetCurLoggerLevel() string {
	return atomicLevel.Level().String()
}

func SetLoggerLevel(lvl string) {
	atomicLevel.SetLevel(getLoggerLevel(lvl))
}

func (conf *LogConfig) parseLog() {
	if conf.Env == "" {
		conf.Env = constant.EnvDev
	}
	if conf.Platform == "" {
		conf.Platform = constant.Platform
	}
	if conf.Service == "" {
		conf.Service = "unkonw"
	}

	if conf.Path == "" {
		conf.Path = "./info.log"
	}

	if conf.MaxSize == 0 {
		conf.MaxSize = maxSize
	}

	if conf.MaxBackups == 0 {
		conf.MaxBackups = maxBackups
	}

	if conf.MaxAge == 0 {
		conf.MaxAge = maxAge
	}

	if conf.LogLevel == "" {
		conf.LogLevel = "info"
	}
}

func Init(conf *LogConfig) {
	conf.parseLog()
	var opts []otelzap.Option
	opts = append(opts,
		otelzap.WithMinLevel(getLoggerLevel(conf.LogLevel)),
		otelzap.WithStackTrace(true))
	switch conf.Env {
	case constant.EnvDev: // 开发环境
		InitStdOutCtxLogger(conf, opts...)
	case constant.EnvProd, constant.EnvUat, constant.EnvTest: // 测试生产环境
		InitCtxLogger(conf, opts...)
	}
}

func Debugf(ctx context.Context, format string, v ...any) {
	if !IsInitialized() {
		fmt.Printf(format, v...)
		return
	}

	CtxLogger(ctx).Debug(fmt.Sprintf(format, v...))
}

func Infof(ctx context.Context, format string, v ...any) {
	if !IsInitialized() {
		fmt.Printf(format+"\n", v...)
		return
	}

	CtxLogger(ctx).Info(fmt.Sprintf(format, v...))
}

func Warnf(ctx context.Context, format string, v ...any) {
	if !IsInitialized() {
		fmt.Printf(format+"\n", v...)
		return
	}

	CtxLogger(ctx).Warn(fmt.Sprintf(format, v...))
}

func Errorf(ctx context.Context, format string, v ...any) {
	if !IsInitialized() {
		fmt.Printf(format+"\n", v...)
		return
	}

	CtxLogger(ctx).Error(fmt.Sprintf(format, v...))
}

func Fatalf(ctx context.Context, format string, v ...any) {
	if !IsInitialized() {
		fmt.Printf(format+"\n", v...)
		return
	}

	CtxLogger(ctx).Fatal(fmt.Sprintf(format, v...))
}

func Close() error {
	if !IsInitialized() {
		return nil
	}

	_ = ctxLogger.Sync()
	lumberjackLoggerClose()
	return nil
}

func IsInitialized() bool { return ctxLogger != nil }
