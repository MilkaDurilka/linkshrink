package errors

import (
	"errors"
	"fmt"
)

type UniqueViolationError struct {
	Err error
}

func (e *UniqueViolationError) Error() string {
	return fmt.Sprintf("Ошибка уникальности: %v", e.Err)
}

func IsUniqueViolation(err error) bool {
	var myErr *UniqueViolationError
	return errors.As(err, &myErr)
}
