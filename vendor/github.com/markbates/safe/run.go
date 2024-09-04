package safe

import "fmt"

type RunFn func() error

// Run the function safely knowing that if it panics
// the panic will be caught and returned as an error
func Run(fn RunFn) (err error) {
	defer func() {
		if err != nil {
			return
		}

		r := recover()
		if r == nil {
			return
		}

		switch t := r.(type) {
		case error:
			err = t
		case string:
			err = fmt.Errorf(t)
		default:
			err = fmt.Errorf("%+v", t)
		}
	}()

	return fn()
}
