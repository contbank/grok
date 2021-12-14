package grok

// ErrorMapping ...
type ErrorMapping map[string]error

var (
	// DefaultErrorMapping ...
	DefaultErrorMapping = ErrorMapping{}
)

// Register ...
func (mapping ErrorMapping) Register(k string, v error) {
	mapping[k] = v
}

// Exists ...
func (mapping ErrorMapping) Exists(err string) bool {
	_, has := mapping[err]

	return has
}

// Get ...
func (mapping ErrorMapping) Get(err string) error {
	return mapping[err]
}
