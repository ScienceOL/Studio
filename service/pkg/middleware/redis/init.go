package redis

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"

	"github.com/redis/go-redis/extra/rediscmd/v9"
	r "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumName = "github.com/redis/go-redis/extra/redisotel"
)

type Redis struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Init 初始化redis连接
func initRedis(conf *Redis) (*r.Client, error) {
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	client := r.NewClient(&r.Options{
		Addr:     addr,
		Password: conf.Password,
		DB:       conf.DB,
	})
	client.AddHook(newTracingHook(addr))
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

type config struct {
	// Common options.

	dbSystem string
	attrs    []attribute.KeyValue

	// Tracing options.

	tp     trace.TracerProvider
	tracer trace.Tracer

	dbStmtEnabled bool
	callerEnabled bool

	// Metrics options.

	mp metric.MeterProvider
	// meter metric.Meter

	// poolName string

	closeChan chan struct{}
}

type baseOption interface {
	apply(conf *config)
}

type Option interface {
	baseOption
	tracing()
	metrics()
}

type option func(conf *config)

func (fn option) apply(conf *config) {
	fn(conf)
}

func (fn option) tracing() {}

func (fn option) metrics() {}

func newConfig(opts ...baseOption) *config {
	conf := &config{
		dbSystem: "redis",
		attrs:    []attribute.KeyValue{},

		tp:            otel.GetTracerProvider(),
		mp:            otel.GetMeterProvider(),
		dbStmtEnabled: true,
		callerEnabled: true,
	}

	for _, opt := range opts {
		opt.apply(conf)
	}

	conf.attrs = append(conf.attrs, semconv.DBSystemKey.String(conf.dbSystem))

	return conf
}

func WithDBSystem(dbSystem string) Option {
	return option(func(conf *config) {
		conf.dbSystem = dbSystem
	})
}

// WithAttributes specifies additional attributes to be added to the span.
func WithAttributes(attrs ...attribute.KeyValue) Option {
	return option(func(conf *config) {
		conf.attrs = append(conf.attrs, attrs...)
	})
}

//------------------------------------------------------------------------------

type TracingOption interface {
	baseOption
	tracing()
}

type tracingOption func(conf *config)

var _ TracingOption = (*tracingOption)(nil)

func (fn tracingOption) apply(conf *config) {
	fn(conf)
}

func (fn tracingOption) tracing() {}

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracerProvider(provider trace.TracerProvider) TracingOption {
	return tracingOption(func(conf *config) {
		conf.tp = provider
	})
}

// WithDBStatement tells the tracing hook to log raw redis commands.
func WithDBStatement(on bool) TracingOption {
	return tracingOption(func(conf *config) {
		conf.dbStmtEnabled = on
	})
}

// WithCallerEnabled tells the tracing hook to log the calling function, file and line.
func WithCallerEnabled(on bool) TracingOption {
	return tracingOption(func(conf *config) {
		conf.callerEnabled = on
	})
}

//------------------------------------------------------------------------------

type MetricsOption interface {
	baseOption
	metrics()
}

type metricsOption func(conf *config)

var _ MetricsOption = (*metricsOption)(nil)

func (fn metricsOption) apply(conf *config) {
	fn(conf)
}

func (fn metricsOption) metrics() {}

// WithMeterProvider configures a metric.Meter used to create instruments.
func WithMeterProvider(mp metric.MeterProvider) MetricsOption {
	return metricsOption(func(conf *config) {
		conf.mp = mp
	})
}

func WithCloseChan(closeChan chan struct{}) MetricsOption {
	return metricsOption(func(conf *config) {
		conf.closeChan = closeChan
	})
}

type tracingHook struct {
	conf *config

	spanOpts []trace.SpanStartOption
}

var _ r.Hook = (*tracingHook)(nil)

func newTracingHook(connString string, opts ...TracingOption) *tracingHook {
	baseOpts := make([]baseOption, len(opts))
	for i, opt := range opts {
		baseOpts[i] = opt
	}
	conf := newConfig(baseOpts...)

	if conf.tracer == nil {
		conf.tracer = conf.tp.Tracer(
			instrumName,
			trace.WithInstrumentationVersion("semver:"+r.Version()),
		)
	}
	if connString != "" {
		conf.attrs = append(conf.attrs, semconv.DBConnectionString(connString))
	}

	return &tracingHook{
		conf: conf,

		spanOpts: []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(conf.attrs...),
		},
	}
}

func (th *tracingHook) DialHook(hook r.DialHook) r.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		ctx, span := th.conf.tracer.Start(ctx, "redis.dial", th.spanOpts...)
		defer span.End()

		conn, err := hook(ctx, network, addr)
		if err != nil {
			recordError(span, err)
			return nil, err
		}
		return conn, nil
	}
}

func (th *tracingHook) ProcessHook(hook r.ProcessHook) r.ProcessHook {
	return func(ctx context.Context, cmd r.Cmder) error {

		attrs := make([]attribute.KeyValue, 0, 8)
		if th.conf.callerEnabled {
			fn, file, line := funcFileLine("github.com/redis/go-redis")
			attrs = append(attrs,
				semconv.CodeFunction(fn),
				semconv.CodeFilepath(file),
				semconv.CodeLineNumber(line),
			)
		}

		if th.conf.dbStmtEnabled {
			cmdString := rediscmd.CmdString(cmd)
			attrs = append(attrs, semconv.DBStatement(cmdString))
		}

		opts := th.spanOpts
		opts = append(opts, trace.WithAttributes(attrs...))

		ctx, span := th.conf.tracer.Start(ctx, cmd.FullName(), opts...)
		defer span.End()

		if err := hook(ctx, cmd); err != nil {
			recordError(span, err)
			return err
		}
		return nil
	}
}

func (th *tracingHook) ProcessPipelineHook(
	hook r.ProcessPipelineHook,
) r.ProcessPipelineHook {
	return func(ctx context.Context, cmds []r.Cmder) error {
		attrs := make([]attribute.KeyValue, 0, 8)
		attrs = append(attrs,
			attribute.Int("db.redis.num_cmd", len(cmds)),
		)

		if th.conf.callerEnabled {
			fn, file, line := funcFileLine("github.com/redis/go-redis")
			attrs = append(attrs,
				semconv.CodeFunction(fn),
				semconv.CodeFilepath(file),
				semconv.CodeLineNumber(line),
			)
		}

		summary, cmdsString := rediscmd.CmdsString(cmds)
		if th.conf.dbStmtEnabled {
			attrs = append(attrs, semconv.DBStatement(cmdsString))
		}

		opts := th.spanOpts
		opts = append(opts, trace.WithAttributes(attrs...))

		ctx, span := th.conf.tracer.Start(ctx, "redis.pipeline "+summary, opts...)
		defer span.End()

		if err := hook(ctx, cmds); err != nil {
			recordError(span, err)
			return err
		}
		return nil
	}
}

func recordError(span trace.Span, err error) {
	if err != r.Nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// func formatDBConnString(network, addr string) string {
// 	if network == "tcp" {
// 		network = "redis"
// 	}
// 	return fmt.Sprintf("%s://%s", network, addr)
// }

func funcFileLine(pkg string) (string, string, int) {
	const depth = 16
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	ff := runtime.CallersFrames(pcs[:n])

	var fn, file string
	var line int
	for {
		f, ok := ff.Next()
		if !ok {
			break
		}
		fn, file, line = f.Function, f.File, f.Line
		if !strings.Contains(fn, pkg) {
			break
		}
	}

	if ind := strings.LastIndexByte(fn, '/'); ind != -1 {
		fn = fn[ind+1:]
	}

	return fn, file, line
}

// Database span attributes semantic conventions recommended server address and port
// https://opentelemetry.io/docs/specs/semconv/database/database-spans/#connection-level-attributes
// func addServerAttributes(opts []TracingOption, addr string) []TracingOption {
// 	host, portString, err := net.SplitHostPort(addr)
// 	if err != nil {
// 		return opts
// 	}
//
// 	opts = append(opts, WithAttributes(
// 		semconv.ServerAddress(host),
// 	))
//
// 	// Parse the port string to an integer
// 	port, err := strconv.Atoi(portString)
// 	if err != nil {
// 		return opts
// 	}
//
// 	opts = append(opts, WithAttributes(
// 		semconv.ServerPort(port),
// 	))
//
// 	return opts
// }
