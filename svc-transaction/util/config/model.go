package config

// App config

type App struct {
	Name string `mapstructure:"name"`
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
