package grok

import "strings"

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

// Get ...
func (mapping ErrorMapping) Get(err string) error {
	for key, result := range mapping {
		if strings.Contains(strings.ToLower(err), strings.ToLower(key)) {
			return result
		}
	}
	return nil
}
