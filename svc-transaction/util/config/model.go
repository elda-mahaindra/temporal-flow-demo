package config

// App config

type App struct {
	Name string `mapstructure:"name"`
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DB config

type PostgresPool struct {
	MaxConns int `mapstructure:"max_conns"`
	MinConns int `mapstructure:"min_conns"`
}

type PostgresConfig struct {
	ConnectionString string       `mapstructure:"connection_string"`
	Pool             PostgresPool `mapstructure:"pool"`
}

type DB struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
}

// Temporal config with performance optimization settings

type TemporalWorkerOptions struct {
	MaxConcurrentActivityExecutions  int  `mapstructure:"max_concurrent_activity_executions"`
	MaxConcurrentWorkflowExecutions  int  `mapstructure:"max_concurrent_workflow_executions"`
	MaxConcurrentLocalActivities     int  `mapstructure:"max_concurrent_local_activities"`
	MaxConcurrentActivityTaskPollers int  `mapstructure:"max_concurrent_activity_task_pollers"`
	MaxConcurrentWorkflowTaskPollers int  `mapstructure:"max_concurrent_workflow_task_pollers"`
	EnableSessionWorker              bool `mapstructure:"enable_session_worker"`
}

type Temporal struct {
	HostPort      string                `mapstructure:"host_port"`
	Namespace     string                `mapstructure:"namespace"`
	TaskQueue     string                `mapstructure:"task_queue"`
	WorkerOptions TemporalWorkerOptions `mapstructure:"worker_options"`
}
