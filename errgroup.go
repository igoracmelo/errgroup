package errgroup

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Group struct {
	wg     sync.WaitGroup
	cancel context.CancelFunc
	sem    chan struct{}
	errs   chan error
}

func New() *Group {
	return &Group{
		errs: make(chan error),
	}
}

func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)

	return &Group{
		errs:   make(chan error),
		cancel: cancel,
	}, ctx
}

func (g *Group) SetLimit(limit int) {
	g.sem = make(chan struct{}, limit)
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	if g.sem != nil {
		g.sem <- struct{}{}
	}

	go func() {
		defer g.wg.Done()

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
		if err != nil {
			g.errs <- err
		}
	}()
}

func (g *Group) Wait() error {
	var err error

	go func() {
		g.wg.Wait()
		close(g.errs)
	}()

	for curr := range g.errs {
		if err == nil {
			err = curr
		}
		if g.cancel != nil {
			g.cancel()
		}
	}

	if g.cancel != nil {
		g.cancel()
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
