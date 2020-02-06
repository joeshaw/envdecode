package envdecode

type config struct {
	strict     bool
	require    bool
	nodefaults bool
}

type Option func(cfg *config)

func WithStrictDecoding() Option {
	return func(cfg *config) {
		cfg.strict = true
	}
}

func WithForcedRequirement() Option {
	return func(cfg *config) {
		cfg.require = true
	}
}

func WithoutDefaults() Option {
	return func(cfg *config) {
		cfg.nodefaults = true
	}
}

func newConfig(options ...Option) config {
	cfg := config{}
	for _, option := range options {
		option(&cfg)
	}
	return cfg
}
