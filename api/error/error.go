package error

// The Code should be a const in https://golang.org/pkg/net/http/
type AppError struct {
	Err     error
	Message string
	Code    int
}
