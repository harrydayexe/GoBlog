package config

type ServerOption struct {
	BaseOption

	WithPortFunc func(v *Port)
	WithHostFunc func(v *Host)
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

type Host int

func WithHost(host int) ServerOption {
	return ServerOption{
		WithHostFunc: func(v *Host) {
			// HACK: Silly hack as you can't do &BlogRoot(root) all at once
			x := Host(host)
			v = &x
		},
	}
}
