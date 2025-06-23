package config

// App config

type App struct {
	Name string `mapstructure:"name"`
	Host string `mapstructure:"host"` // Bind address (0.0.0.0 for listening)
	Port int    `mapstructure:"port"`
}

// Temporal config

type Temporal struct {
	HostPort        string                  `mapstructure:"host_port"`
	Namespace       string                  `mapstructure:"namespace"`
	ActivityOptions TemporalActivityOptions `mapstructure:"activity_options"`
}

// TemporalActivityOptions defines configuration for activity execution behavior
type TemporalActivityOptions struct {
	StartToCloseTimeoutSeconds    int                 `mapstructure:"start_to_close_timeout_seconds"`
	HeartbeatTimeoutSeconds       int                 `mapstructure:"heartbeat_timeout_seconds"`
	ScheduleToCloseTimeoutSeconds int                 `mapstructure:"schedule_to_close_timeout_seconds"`
	ScheduleToStartTimeoutSeconds int                 `mapstructure:"schedule_to_start_timeout_seconds"`
	RetryPolicy                   TemporalRetryPolicy `mapstructure:"retry_policy"`
}

// TemporalRetryPolicy defines retry behavior for activities
type TemporalRetryPolicy struct {
	InitialIntervalMs      int      `mapstructure:"initial_interval_ms"`
	BackoffCoefficient     float64  `mapstructure:"backoff_coefficient"`
	MaximumIntervalSeconds int      `mapstructure:"maximum_interval_seconds"`
	MaximumAttempts        int      `mapstructure:"maximum_attempts"`
	NonRetryableErrorTypes []string `mapstructure:"non_retryable_error_types"`
}
