package logger

import "go.uber.org/zap/zapcore"

type Log struct {
	Path               string
	MaxSize            int  // 单文件最大容量(单位MB)
	MaxBackups         int  // 保留旧文件的最大数量
	MaxAge             int  // 旧文件最多保存几天
	Compress           bool // 是否压缩/归档旧文件
	OpenTraceInfoLevel int  // 开启info级别的trace日志
	TraceLogMinLevel   zapcore.Level
}

type ServiceEnv struct {
	Platform string
	Service  string
	Env      string
}

func NewSafeLog(l *Log) *Log {
	if l == nil {
		return NewLog()
	}
	opts := make([]Option, 0)
	if l.Path != "" {
		opts = append(opts, WithPath(l.Path))
	}
	if l.MaxSize > 0 {
		opts = append(opts, WithMaxSize(l.MaxSize))
	}
	if l.MaxBackups > 0 {
		opts = append(opts, WithMaxBackups(l.MaxBackups))
	}
	if l.MaxAge > 0 {
		opts = append(opts, WithMaxAge(l.MaxAge))
	}
	if l.Compress {
		opts = append(opts, WithCompress(l.Compress))
	}
	if l.OpenTraceInfoLevel > 0 {
		opts = append(opts, WithTraceLogMinLevel(zapcore.InfoLevel))
	}
	if l.TraceLogMinLevel > 0 {
		opts = append(opts, WithTraceLogMinLevel(l.TraceLogMinLevel))
	}
	return NewLog(opts...)
}

func NewLog(opts ...Option) *Log {
	l := &Log{
		Path:             "/data/logs/zap.log",
		MaxSize:          50,
		MaxBackups:       10,
		MaxAge:           7,
		Compress:         false,
		TraceLogMinLevel: zapcore.WarnLevel,
	}
	for _, f := range opts {
		f(l)
	}
	return l
}

type Option func(*Log)

func WithPath(path string) Option {
	return func(l *Log) {
		l.Path = path
	}
}

func WithMaxSize(maxSize int) Option {
	return func(l *Log) {
		l.MaxSize = maxSize
	}
}

func WithMaxBackups(maxBackups int) Option {
	return func(l *Log) {
		l.MaxBackups = maxBackups
	}
}

func WithMaxAge(maxAge int) Option {
	return func(l *Log) {
		l.MaxAge = maxAge
	}
}

func WithCompress(compress bool) Option {
	return func(l *Log) {
		l.Compress = compress
	}
}

func WithTraceLogMinLevel(level zapcore.Level) Option {
	return func(l *Log) {
		l.TraceLogMinLevel = level
	}
}
