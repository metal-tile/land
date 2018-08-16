package recoverable

import "fmt"

type errRecovered struct {
	value interface{}
}

func (err *errRecovered) Error() string {
	return fmt.Sprintf("panic with %s", err.value)
}

func (err *errRecovered) RecoveredValue() interface{} {
	return err.value
}

// RecoveredValue returns recovered value from the error.
// If the error implements bellow interface,
// Recoveredvalue returns the recovered value and true.
//     interface {
//          RecoveredValue() interface{}
//     }
func RecoveredValue(err error) (interface{}, bool) {
	rerr, ok := err.(interface {
		RecoveredValue() interface{}
	})

	if !ok {
		return nil, false
	}
	return rerr.RecoveredValue(), true
}

// Func converts the given function to a function
// which returns an error when a panic
// have occured in the given function.
// The recovered value can get from the error with RecoveredValue.
func Func(f func()) func() error {
	return func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = &errRecovered{value: r}
			}
		}()
		f()
		return
	}
}
