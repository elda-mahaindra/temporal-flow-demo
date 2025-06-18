package config

// App config

type App struct {
	Name string `mapstructure:"name"`
	Host string `mapstructure:"host"` // Bind address (0.0.0.0 for listening)
	Port int    `mapstructure:"port"`
}

// Flowngine config

type Flowngine struct {
	Name string `mapstructure:"name"`
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}
