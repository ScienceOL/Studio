package trace

import (
	"context"
	"fmt"
)

type InitConfig struct {
	ServiceName     string
	Version         string
	TraceEndpoint   string
	MetricEndpoint  string
	TraceProject    string
	TraceInstanceID string
	TraceAK         string
	TraceSK         string
}

var client *Config

// 不做强制要求
func InitTrace(_ context.Context, conf *InitConfig) {
	if conf.ServiceName == "" ||
		conf.TraceEndpoint == "" ||
		conf.MetricEndpoint == "" ||
		conf.TraceInstanceID == "" {
		fmt.Println("not init trace")
		return
	}

	var err error
	client, err = NewConfig(WithServiceName(conf.ServiceName),
		WithServiceVersion(conf.Version),
		WithTraceExporterEndpoint(conf.TraceEndpoint),
		WithMetricExporterEndpoint(conf.MetricEndpoint),
		WithSLSConfig(conf.TraceProject, conf.TraceInstanceID, conf.TraceAK, conf.TraceSK))
	if err != nil {
		fmt.Printf("init trace config err: %s\n", err.Error())
		client = nil
	}

	err = Start(client)
	if err != nil {
		fmt.Printf("init trace provider err: %s\n", err.Error())
		client = nil
	}
}

func CloseTrace() {
	if client != nil {
		Shutdown(client)
	}
}
