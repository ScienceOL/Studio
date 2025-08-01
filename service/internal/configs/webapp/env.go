package webapp

// 规划环境变量
type Database struct {
	Host     string `mapstructure:"DATABASE_HOST" default:"localhost"`
	Port     int    `mapstructure:"DATABASE_PORT" default:"5432"`
	Name     string `mapstructure:"DATABASE_NAME" default:"studio"`
	User     string `mapstructure:"DATABASE_USER" default:"postgres"`
	Password string `mapstructure:"DATABASE_PASSWORD" default:"studio"`
}

type Redis struct {
	Host     string `mapstructure:"REDIS_HOST" default:"127.0.0.1"`
	Port     int    `mapstructure:"REDIS_PORT" default:"6379"`
	User     string `mapstructure:"REDIS_USER" `
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB" default:"0"`
}

type Server struct {
	Platform  string `mapstructure:"PLATFORM" default:"sciol"` // linux、darwin、windows
	Service   string `mapstructure:"SERVICE" default:"studio"` // api、schedule
	SecretKey string `mapstructure:"SECRET_KEY"`
	Port      int    `mapstructure:"SERVER_PORT" default:"48197"`
	Env       string `mapstructure:"ENV" default:"dev"`
}

type OAuth2 struct {
	ClientID     string   `mapstructure:"OAUTH2_CLIENT_ID" default:"a387a4892ee19b1a2249"`
	ClientSecret string   `mapstructure:"OAUTH2_CLIENT_SECRET" default:"f3167664b2c58bca53b04c61807a97db"`
	Scopes       []string `mapstructure:"OAUTH2_SCOPES" default:"[\"read\",\"write\",\"offline_access\"]"`
	TokenURL     string   `mapstructure:"OAUTH2_TOKEN_URL" default:"http://localhost:8000/api/login/oauth/access_token"`
	AuthURL      string   `mapstructure:"OAUTH2_AUTH_URL" default:"http://localhost:8000/login/oauth/authorize"`
	RedirectURL  string   `mapstructure:"OAUTH2_REDIRECT_URL" default:"http://localhost:48197/api/auth/callback/casdoor"`
	UserInfoURL  string   `mapstructure:"OAUTH2_USERINFO_URL" default:"http://localhost:8000/api/get-account"`
}

type Log struct {
	LogPath  string `mapstructure:"LOG_PATH" default:"./info.log"`
	LogLevel string `mapstructure:"LOG_LEVEL" default:"info"`
}

type MQTT struct {
	AccessKey  string `mapstructure:"MQTT_ACCESS_KEY" default:""`
	SecretKey  string `mapstructure:"MQTT_SECRET_KEY" default:""`
	InstanceID string `mapstructure:"MQTT_INSTANCEID" default:"mqtt-cn-kjp48jwyv01"`
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
	NamespaceID string `mapstructure:"NACOS_NAMESPACE_ID" default:"public"`
	AccessKey   string `mapstructure:"NACOS_ACCESS_KEY" default:""`
	SecretKey   string `mapstructure:"NACOS_SECRET_KEY" default:""`
	User        string `mapstructure:"NACOS_USER" default:"nacos"`
	Password    string `mapstructure:"NACOS_PASSWORD" default:"nacos"`
	Port        uint64 `mapstructure:"NACOS_PORT" default:"8848"`
	RegionID    string `mapstructure:"NACOS_REGION_ID" default:""`
	DataID      string `mapstructure:"NACOS_DATA_ID" default:"studio-config"`
	Group       string `mapstructure:"NACOS_GROUP" default:"DEFAULT_GROUP"`
	NeedWatch   bool   `mapstructure:"NACOS_NEED_WATCH" default:"true"`
}
