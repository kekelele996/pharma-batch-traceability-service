package platform

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrValidation = errors.New("validation failed")
)

func WrapNotFound(what string) error { return fmt.Errorf("%s: %w", what, ErrNotFound) }
func WrapConflict(what string) error { return fmt.Errorf("%s: %w", what, ErrConflict) }
func WrapValidation(what string) error { return fmt.Errorf("%s: %w", what, ErrValidation) }

func IsNotFound(err error) bool   { return errors.Is(err, ErrNotFound) }
func IsConflict(err error) bool   { return errors.Is(err, ErrConflict) }
func IsValidation(err error) bool { return errors.Is(err, ErrValidation) }
