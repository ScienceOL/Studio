package schedule

// 规划尽快使用 nacos 获取动态配置

type DagTask struct {
	RetryCount int `yaml:"retry_count" mapstructure:"retry_count" default:"120"`
	Interval   int `yaml:"interval" mapstructure:"interval" default:"1"`
}

type DynamicConfig struct {
	DagTask DagTask `yaml:"dag_task"`
}
