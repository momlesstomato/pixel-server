package logging

// Config defines logger output and verbosity behavior.
type Config struct {
	// Format selects structured output format: json or console.
	Format string `mapstructure:"format" default:"console"`
	// Level selects threshold level: debug, info, warn, or error.
	Level string `mapstructure:"level" default:"info"`
}
