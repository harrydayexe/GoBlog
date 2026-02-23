package config

type ServerOption struct {
	WithPortFunc func(v *Port)
}

type Port int

func WithPort(port int) ServerOption {
	return ServerOption{
		WithPortFunc: func(v *Port) {
			// HACK: Silly hack as you can't do &BlogRoot(root) all at once
			x := Port(port)
			v = &x
		},
	}
}
