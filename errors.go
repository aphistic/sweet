package sweet

import "errors"

var (
	errUnsupportedVersion = errors.New("This version of Go is unsupported")
	errUnknownResponse    = errors.New("Unknown response from runtime")
	errDeprecated         = errors.New("Deprecated")
	errUnsupportedMethod  = errors.New("Unsupported method")
	errInvalidValue       = errors.New("Invalid value")
)
