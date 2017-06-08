package error

import (
	"errors"
	"fmt"
	"net/http"
)

func (e *AppError) Error() string {
	return fmt.Sprintf("%v: %v", e.Message, e.Err)
}

func newAppError(err error, msg string, code int) AppError {
	e := errors.New(fmt.Sprintf("%v: %v", msg, err))
	return AppError{e, msg, http.StatusNotFound}
}

func ExampleError() {
	err := func() error {
		return errors.New("Return a error just for test.")
	}
	// In a model, after we catch an error, reform it and return AppError.
	appErr := AppError{err(), "Reason Message", http.StatusNotFound}
	fmt.Println(fmt.Sprintf("%v", appErr))
	fmt.Println(appErr.Error(), appErr.Code)

	appErr = newAppError(err(), "Reason Message", http.StatusNotFound)
	fmt.Println(appErr.Err, appErr.Code)
	// Output:
	// {Return a error just for test. Reason Message 404}
	// Reason Message: Return a error just for test. 404
	// Reason Message: Return a error just for test. 404
}
