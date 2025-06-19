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

// Temporal config

type Temporal struct {
	HostPort  string `mapstructure:"host_port"`
	Namespace string `mapstructure:"namespace"`
	TaskQueue string `mapstructure:"task_queue"`
}
