package errors

import "fmt"

var (
	ErrUnauthorized = fmt.Errorf("unauthorized — invalid or expired credentials")
	ErrForbidden    = fmt.Errorf("forbidden — insufficient permissions")
)
