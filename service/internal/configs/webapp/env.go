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

type MQTT struct {
	AccessKey  string `mapstructure:"MQTT_ACCESS_KEY" default:""`
	SecretKey  string `mapstructure:"MQTT_SECRET_KEY" default:""`
	InstanceId string `mapstructure:"MQTT_INSTANCEID" default:"mqtt-cn-kjp48jwyv01"`
	Domain     string `mapstructure:"MQTT_DOMAIN" default:"mqtt-cn-kjp48jwyv01-server.mqtt.aliyuncs.com"`
	Port       int16  `mapstructure:"MQTT_PORT" default:"5672"`
	Topic      string `mapstructure:"MQTT_TOPIC" default:"labs"`
	Gid        string `mapstructure:"MQTT_GID" default:"GID_share_test"`
}

type Trace struct {
	Version         string `mapstructure:"TRACE_VERSION" default:"0.0.1"`
	TraceEndpoint   string `mapstructure:"TRACE_TRACEENDPOINT" default:""`
	MetricEndpoint  string `mapstructure:"TRACE_METRICENDPOINT" default:""`
	TraceProject    string `mapstructure:"TRACE_TRACEPROJECT" default:""`
	TraceInstanceID string `mapstructure:"TRACE_TRACEINSTANCEID" default:""`
	TraceAK         string `mapstructure:"TRACE_TRACEAK" default:""`
	TraceSK         string `mapstructure:"TRACE_TRACESK" default:""`
}

type Nacos struct {
	Endpoint    string `mapstructure:"NACOS_ENDPOINT" default:"127.0.0.1"`
	ContextPath string `mapstructure:"NACOS_CONTEXT_PATH" default:"/nacos"`
	NamespaceID string `mapstructure:"NACOS_NAMESPACE_ID" default:""`
	AccessKey   string `mapstructure:"NACOS_ACCESS_KEY" default:""`
	SecretKey   string `mapstructure:"NACOS_SECRET_KEY" default:""`
	User        string `mapstructure:"NACOS_USER" default:"nacos"`
	Password    string `mapstructure:"NACOS_PASSWORD" default:"nacos"`
	Port        int    `mapstructure:"NACOS_PORT" default:"8848"`
	RegionID    string `mapstructure:"NACOS_REGION_ID" default:""`
	DataID      string `mapstructure:"NACOS_DATA_ID"`
	Group       string `mapstructure:"NACOS_GROUP" default:"DEFAULT_GROUP"`
	NeedWatch   bool   `mapstructure:"NACOS_NEED_WATCH" default:"true"`
}
