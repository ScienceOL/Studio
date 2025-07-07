package webapp

import (
	"fmt"
	"os"

	"github.com/creasty/defaults"
)

type WebGlobalConfig struct {
	Database Database `mapstructure:",squash"`
	Redis    Redis    `mapstructure:",squash"`
	Server   Server   `mapstructure:",squash"`
	Log      Log      `mapstructure:",squash"`
	MQTT     MQTT     `mapstructure:",squash"`
	Trace    Trace    `mapstructure:",squash"`
}

var config = &WebGlobalConfig{}

func init() {
	if err := defaults.Set(config); err != nil {
		fmt.Printf("set default err: %+v", err)
		os.Exit(1)
	}
}

func Config() *WebGlobalConfig {
	return config
}
