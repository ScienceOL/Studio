package webapp

// 规划环境变量
type Database struct {
	Host     string `mapstructure:"DB_HOST" default:""`
	Port     int    `mapstructure:"DB_PORT" default:"5432"`
	User     string `mapstructure:"DB_USER" default:"postgres"`
	Password string `mapstructure:"DB_PASSWORD" default:""`
	Name     string `mapstructure:"DB_NAME" default:"protium"`
}

type Redis struct {
	Host     string `mapstructure:"REDIS_HOST" `
	Port     int    `mapstructure:"REDIS_PORT"`
	User     string `mapstructure:"REDIS_USER"`
	Password string `mapstructure:"REDIS_PASSWORD"`
}

type Server struct {
	Port     int    `mapstructure:"SERVER_PORT"`
	Platform string `mapstructure:"PLATFORM"` // uni-lab
	Service  string `mapstructure:"SERVICE"`  // api、schedule
	Env      string `mapstructure:"ENV"`
}

type Log struct {
	LogPath string `mapstructure:"LOG_PATH"`
}
