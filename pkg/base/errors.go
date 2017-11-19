package base

import "fmt"

const (
    ErrCodeNotImplement    = 1
    ErrCodeNotInvalidParam = 2
    ErrCodeNotFound        = 10
)

func NewError(code int) error {
    return revolverError{Code: code}
}

type revolverError struct {
    Code int
}

func (e revolverError) Error() string {
    return fmt.Sprintf("err code: %d", e.Code)
}
