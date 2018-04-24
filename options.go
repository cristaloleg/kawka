package kawka

// Option ...
type Option func(*Kawka) error

// WithBrokers ...
func WithBrokers(brokers []string) Option {
	return func(wk *Kawka) error {
		wk.brokers = brokers
		return nil
	}
}

// WithHandler ...
func WithHandler(handler MessageHandler) Option {
	return func(wk *Kawka) error {
		wk.handler = handler
		return nil
	}
}

// WithPort ...
func WithPort(port int) Option {
	return func(wk *Kawka) error {
		wk.port = port
		return nil
	}
}

// WithTSL ...
func WithTSL(certFile, keyFile string) Option {
	return func(wk *Kawka) error {
		return nil
	}
}
