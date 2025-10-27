package config

import (
	"fmt"
	"os"

	"github.com/creasty/defaults"
)

type GlobalConfig struct {
	Database      Database `mapstructure:",squash"`
	Redis         Redis    `mapstructure:",squash"`
	Server        Server   `mapstructure:",squash"`
	OAuth2        OAuth2   `mapstructure:",squash"`
	Log           Log      `mapstructure:",squash"`
	// Trace         Trace    `mapstructure:",squash"`
	// Nacos         Nacos    `mapstructure:",squash"`
	Job           Job      `mapstructure:",squash"`
	RPC           RPC      `mapstructure:",squash"`
	Auth          Auth     `mapstructure:",squash"`
	Storage 	  Storage  `mapstructure:",squash"`
	// dynamicConfig *DynamicConfig
}



// func (g *GlobalConfig) SetDynamic(d *DynamicConfig) {
// 	g.dynamicConfig = d
// }

// func (g *GlobalConfig) Dynamic() *DynamicConfig {
// 	return g.dynamicConfig
// }

var config = &GlobalConfig{
	// dynamicConfig: &DynamicConfig{},
}

func init() {
	// 初始化 tag default 值
	if err := defaults.Set(config); err != nil {
		fmt.Printf("set default err: %+v", err)
		os.Exit(1)
	}
}

func Global() *GlobalConfig {
	return config
}
