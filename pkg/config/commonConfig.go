package config

type CommonConfig struct {
	// When true, output should contain only the raw HTML and not be inserted
	// into a template
	RawOutput bool
}

// CommonOption is a function which modifies a CommonConfig.
// Options are used to configure optional parameters when creating a new config.
type CommonOption func(*CommonConfig)

// WithRawOutput sets the config to only generate the raw HTML for each post
// without inserting it into a template.
func WithRawOutput() CommonOption {
	return func(gc *CommonConfig) {
		gc.RawOutput = true
	}
}
