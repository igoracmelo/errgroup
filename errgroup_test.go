package errgroup

import (
	"io"
	"runtime"
	"testing"
)

func TestSingleGoFuncReturningNil(t *testing.T) {
	g := New()

	g.Go(func() error {
		return nil
	})

	err := g.Wait()

	if err != nil {
		t.Error(err)
	}
}

func TestSingleGoFuncReturningError(t *testing.T) {
	g := New()

	g.Go(func() error {
		return io.EOF
	})

	err := g.Wait()

	if err != io.EOF {
		t.Errorf("err - want %v, got %v", io.EOF, err)
	}
}

func TestTwoGoFuncReturningErrAndNil(t *testing.T) {
	g := New()

	g.Go(func() error {
		return nil
	})

	g.Go(func() error {
		return io.EOF
	})

	err := g.Wait()

	if err != io.EOF {
		t.Errorf("err - want %v, got %v", io.EOF, err)
	}
}

func TestAllGoFuncsAreProperlyBeingRunned(t *testing.T) {
	g := New()

	ch := make(chan int)

	g.Go(func() error {
		ch <- 1
		return nil
	})

	g.Go(func() error {
		ch <- 2
		return io.EOF
	})

	g.Go(func() error {
		ch <- 4
		return nil
	})

	sum := <-ch + <-ch + <-ch
	err := g.Wait()

	if err != io.EOF {
		t.Errorf("err - want %v, got %v", io.EOF, err)
	}

	if sum != 7 {
		t.Errorf("sum - want 7, got %d", sum)
	}
}

func TestPanicIsTreatedAsError(t *testing.T) {
	g := New()

	g.Go(func() error {
		return nil
	})

	g.Go(func() error {
		return nil
	})

	g.Go(func() error {
		var x io.Closer
		x.Close()
		return nil
	})

	err := g.Wait()
	_, ok := err.(runtime.Error)
	if !ok {
		t.Errorf("err type - want: runtime.Error, got: %v", err)
	}
}

func TestLimit1ShouldCountCorrectly(t *testing.T) {
	g := New()
	g.SetLimit(1)
	count := 0

	for i := 1; i <= 10e5; i++ {
		g.Go(func() error {
			count++
			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		t.Fatal(err)
	}

	if count != 10e5 {
		t.Errorf("count - want: 10e9, got: %d", count)
	}
}
