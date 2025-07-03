package webapp

type WebGlobalConfig struct {
	Database Database `mapstructure:",squash"`
	Redis    Redis    `mapstructure:",squash"`
	Server   Server   `mapstructure:",squash"`
	Log      Log      `mapstructure:",squash"`
}

var config = &WebGlobalConfig{}

func Config() *WebGlobalConfig {
	return config
}
