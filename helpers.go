package errs

import (
	"errors"
)

func IsAny(err error, references ...error) bool {
	for _, reference := range references {
		if errors.Is(err, reference) {
			return true
		}
	}
	return false
}
