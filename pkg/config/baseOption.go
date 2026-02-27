package config

// BaseOption represents a configuration option that can be applied to
// many different instances during construction.
//
// Options use the functional options pattern, where each option function
// returns an BaseOption struct containing one or more function pointers that
// modify specific configuration fields.
//
// This type should not be constructed directly by users. Instead, use the
// provided option functions like WithBlogRoot().
type BaseOption struct {
	WithBlogRootFunc func(v *BlogRoot)
}

// BlogRoot is a configuration type that holds the blog's root path
//
// This type is typically embedded in generator configuration structs
// and should be set using the WithBlogRoot() option function.
type BlogRoot string

// WithBlogRoot returns an Option that sets the blog's root path.
//
// The blog root is used in generated HTML pages and templates.
//
// Example usage:
//
//	gen := generator.New(fsys, renderer, config.WithBlogRoot("/blog/"))
func WithBlogRoot(root string) BaseOption {
	return BaseOption{
		WithBlogRootFunc: func(v *BlogRoot) {
			// HACK: Silly hack as you can't do &BlogRoot(root) all at once
			x := BlogRoot(root)
			v = &x
		},
	}
}

func (o BlogRoot) AsOption() BaseOption {
	return WithBlogRoot(string(o))
}
