package manifest

type deleteOptions struct {
	Wait bool
}

type DeleteOptionFunc func(c *deleteOptions)

func newDeleteOptions(opts ...DeleteOptionFunc) *deleteOptions {
	options := &deleteOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return options
}

func WaitForDelete() DeleteOptionFunc {
	return func(c *deleteOptions) {
		c.Wait = true
	}
}
