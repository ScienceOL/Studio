package webapp

import (
	"fmt"
	"os"

	"github.com/creasty/defaults"
)

type WebGlobalConfig struct {
	Database      Database `mapstructure:",squash"`
	Redis         Redis    `mapstructure:",squash"`
	Server        Server   `mapstructure:",squash"`
	OAuth2        OAuth2   `mapstructure:",squash"`
	Log           Log      `mapstructure:",squash"`
	MQTT          MQTT     `mapstructure:",squash"`
	Trace         Trace    `mapstructure:",squash"`
	Nacos         Nacos    `mapstructure:",squash"`
	Job           Job      `mapstructure:",squash"`
	DynamicConfig *DynamicConfig
}

var config = &WebGlobalConfig{}

func init() {
	// 初始化 tag default 值
	if err := defaults.Set(config); err != nil {
		fmt.Printf("set default err: %+v", err)
		os.Exit(1)
	}
}

func Config() *WebGlobalConfig {
	return config
}
