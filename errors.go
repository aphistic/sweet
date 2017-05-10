package sweet

import "errors"

var (
	errUnsupportedVersion = errors.New("This version of Go is unsupported")
	errUnknownResponse    = errors.New("Unknown response from runtime")
)
