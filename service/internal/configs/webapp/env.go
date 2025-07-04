package webapp

// 规划环境变量
type Database struct {
	Host     string `mapstructure:"DB_HOST" default:"localhost"`
	Port     int    `mapstructure:"DB_PORT" default:"5432"`
	User     string `mapstructure:"DB_USER" default:"postgres"`
	Password string `mapstructure:"DB_PASSWORD" default:"protium"`
	Name     string `mapstructure:"DB_NAME" default:"please_change_me"`
}

type Redis struct {
	Host     string `mapstructure:"REDIS_HOST" default:"127.0.0.1"`
	Port     int    `mapstructure:"REDIS_PORT" default:"6379"`
	User     string `mapstructure:"REDIS_USER" `
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

type Server struct {
	Port     int    `mapstructure:"SERVER_PORT" default:"48197"`
	Platform string `mapstructure:"PLATFORM" default:"uni-lab"` // uni-lab
	Service  string `mapstructure:"SERVICE" default:"api"`      // api、schedule
	Env      string `mapstructure:"ENV" default:"dev"`
}

type Log struct {
	LogPath  string `mapstructure:"LOG_PATH" default:"./info.log"`
	LogLevel string `mapstructure:"LOG_LEVEL" default:"info"` // debug info warn error dpanic panic fatal
}
