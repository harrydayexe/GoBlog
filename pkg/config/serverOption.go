package config

type BaseServerOption struct {
	BaseOption

	WithPortFunc func(v *Port)
	WithHostFunc func(v *Host)
}

type Port int

func WithPort(port int) BaseServerOption {
	return BaseServerOption{
		WithPortFunc: func(v *Port) {
			// HACK: Silly hack as you can't do &BlogRoot(root) all at once
			x := Port(port)
			v = &x
		},
	}
}

type Host string

func WithHost(host string) BaseServerOption {
	return BaseServerOption{
		WithHostFunc: func(v *Host) {
			// HACK: Silly hack as you can't do &BlogRoot(root) all at once
			x := Host(host)
			v = &x
		},
	}
}
