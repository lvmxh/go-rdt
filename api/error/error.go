package error

import (
	"fmt"
	"net/http"
)

// The Code should be a const in https://golang.org/pkg/net/http/
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e AppError) Error() string {
	if e.Message != "" && e.Err != nil {
		return fmt.Sprintf("%s. Details: %v\n", e.Message, e.Err)
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func AppErrorf(code int, format string, a ...interface{}) *AppError {
	msg := fmt.Sprintf(format, a...)
	ae := AppError{code, msg, nil}
	return &ae
}

func NewAppError(value ...interface{}) *AppError {
	ae := AppError{}
	for i, val := range value {
		if i >= 3 {
			break
		}
		switch v := val.(type) {
		case int:
			ae.Code = v
		case string:
			ae.Message = v
		case error:
			ae.Err = v
		default:
			ae.Message = "Unknown AppError type!"
		}
	}
	if ae.Code == 0 {
		ae.Code = http.StatusInternalServerError
	}
	return &ae
}
