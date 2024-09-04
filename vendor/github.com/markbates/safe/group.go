package safe

import (
	"fmt"

	"golang.org/x/sync/errgroup"
)

type Group struct {
	wg errgroup.Group
}

func (sg *Group) Go(fn RunFn) {
	if sg == nil {
		return
	}

	sg.wg.Go(func() error {
		return Run(fn)
	})
}

func (sg *Group) Wait() error {
	if sg == nil {
		return fmt.Errorf("safe group is nil")
	}

	return sg.wg.Wait()
}
