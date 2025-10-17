package logger

import (
	"context"
	"os"
	"sync"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	once             sync.Once
	ctxLogger        *otelzap.Logger
	lumberjackLogger *lumberjack.Logger
	atomicLevel      = zap.NewAtomicLevel()
	levelMap         = map[string]zapcore.Level{
		"debug":  zapcore.DebugLevel,
		"info":   zapcore.InfoLevel,
		"warn":   zapcore.WarnLevel,
		"error":  zapcore.ErrorLevel,
		"dpanic": zapcore.DPanicLevel,
		"panic":  zapcore.PanicLevel,
		"fatal":  zapcore.FatalLevel,
	}
)

func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

func BaseLogger() *otelzap.Logger {
	return ctxLogger
}

// OtelLogger ensures that the caller does not forget to pass the context.
func CtxLogger(ctx context.Context) otelzap.LoggerWithCtx {
	return ctxLogger.Ctx(ctx)
}

func InitStdOutCtxLogger(conf *LogConfig, opts ...otelzap.Option) {
	once.Do(func() {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.Lock(zapcore.AddSync(os.Stdout)),
			zapcore.DebugLevel,
		)
		l := zap.New(core,
			// zap.Fields(zap.String("platform", conf.Platform),
			// 	zap.String("service", conf.Service)),
			zap.WithCaller(true),
			zap.AddCallerSkip(1),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
		ctxLogger = otelzap.New(l, opts...)
	})
}

func InitCtxLogger(conf *LogConfig, opts ...otelzap.Option) {
	once.Do(func() {
		ws := getLumberjackConfig(conf)
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(ws),
			atomicLevel,
		)
		l := zap.New(core,
			zap.Fields(zap.String("platform", conf.Platform),
				zap.String("service", conf.Service)),
			zap.WithCaller(true),
			zap.AddCallerSkip(1),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
		ctxLogger = otelzap.New(l, opts...)
	})
}

// 获取文件切割和归档配置信息
func getLumberjackConfig(conf *LogConfig) zapcore.WriteSyncer {
	lumberjackLogger = &lumberjack.Logger{
		Filename:   conf.Path,       // 日志文件
		MaxSize:    conf.MaxSize,    // 单文件最大容量(单位MB)
		MaxBackups: conf.MaxBackups, // 保留旧文件的最大数量
		MaxAge:     conf.MaxAge,     // 旧文件最多保存几天
		LocalTime:  true,
		Compress:   conf.Compress, // 是否压缩/归档旧文件
	}
	return zapcore.AddSync(lumberjackLogger)
}

func lumberjackLoggerClose() {
	if lumberjackLogger != nil {
		_ = lumberjackLogger.Close()
	}
}
