package errgroup

import (
	"errors"
	"fmt"
)

type Group struct {
	sem   chan struct{}
	count int
	errs  chan error
}

func New() *Group {
	return &Group{
		errs: make(chan error),
	}
}

func (g *Group) SetLimit(limit int) {
	g.sem = make(chan struct{}, limit)
}

func (g *Group) Go(f func() error) {
	if g.sem != nil {
		g.sem <- struct{}{}
	}
	g.count++

	go func() {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch x := r.(type) {
				case string:
					err = errors.New(x)
				case error:
					err = x
				default:
					err = fmt.Errorf("%v", x)
				}

				g.errs <- err
			}
		}()

		err := f()
		if g.sem != nil {
			<-g.sem
		}
		g.errs <- err
	}()
}

func (g *Group) Wait() error {
	var err error

	for i := 0; i < g.count; i++ {
		curr := <-g.errs
		if curr != nil && err == nil {
			err = curr
		}
	}

	return err
}

func (g *Group) WaitAll() []error {
	errs := []error{}

	for err := range g.errs {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
