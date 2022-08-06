package promise

import (
	"sync"
)

type Promise[T any] struct {
	isResolved bool
	value      *T
	chs        []chan T
	resolveWg  *sync.WaitGroup
	awaitWg    *sync.WaitGroup
	wmx        *sync.Mutex
}

func New[T any](cb func() T) *Promise[T] {
	p := &Promise[T]{
		false,
		nil,
		make([]chan T, 0, 1),
		&sync.WaitGroup{},
		&sync.WaitGroup{},
		&sync.Mutex{},
	}

	go p.run(cb)

	return p
}

func Resolve[T any](value T) *Promise[T] {
	return &Promise[T]{
		true,
		&value,
		make([]chan T, 0, 1),
		&sync.WaitGroup{},
		&sync.WaitGroup{},
		&sync.Mutex{},
	}
}

func FromChannel[T any](ch <-chan T) *Promise[T] {
	p := New[T](func() T {
		return <-ch
	})
	return p
}

func All[T any](ps ...*Promise[T]) *Promise[[]T] {
	if len(ps) == 0 {
		// An empty array results in an empty array
		return Resolve(make([]T, 0, 0))
	}
	return New[[]T](func() []T {
		amount := len(ps)
		r := make([]T, amount, amount)

		var wg sync.WaitGroup
		wg.Add(amount)

		for i, p := range ps {
			go (func(p *Promise[T], i int) {
				r[i] = p.Await()
				wg.Done()
			})(p, i)
		}

		wg.Wait()
		return r
	})
}

func Race[T any](ps ...*Promise[T]) *Promise[T] {
	ret := make(chan T)

	amount := len(ps)
	signal := make([]chan struct{}, amount, amount)

	for i, p := range ps {
		go (func(p *Promise[T], i int) {
			select {
			case value := <-p.Chan():
				ret <- value

				// Signal others to stop listening
				for _, sig := range signal {
					sig <- struct{}{}
				}

			case <-signal[i]:
				// Do nothing just exit function
			}
		})(p, i)
	}
	return FromChannel(ret)
}
