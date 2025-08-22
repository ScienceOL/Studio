package schedule

import (
	"fmt"
	"os"

	"github.com/creasty/defaults"
)

type ScheduleGlobalConfig struct {
	Database      Database       `mapstructure:",squash"`
	Redis         Redis          `mapstructure:",squash"`
	Server        Server         `mapstructure:",squash"`
	OAuth2        OAuth2         `mapstructure:",squash"`
	Log           Log            `mapstructure:",squash"`
	Trace         Trace          `mapstructure:",squash"`
	Nacos         Nacos          `mapstructure:",squash"`
	Job           Job            `mapstructure:",squash"`
	DynamicConfig *DynamicConfig 
}

var config = &ScheduleGlobalConfig{}

func init() {
	// 初始化 tag default 值
	if err := defaults.Set(config); err != nil {
		fmt.Printf("set default err: %+v", err)
		os.Exit(1)
	}
}

func Config() *ScheduleGlobalConfig {
	return config
}
