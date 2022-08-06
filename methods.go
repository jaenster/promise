package promise

func (p *Promise[T]) createWaitCh() chan T {
	p.awaitWg.Add(1)
	p.wmx.Lock()
	defer p.awaitWg.Done()
	defer p.wmx.Unlock()

	ch := make(chan T)
	p.chs = append(p.chs, ch)
	return ch
}

func (p *Promise[T]) Await() T {
	// If resolving, wait for it
	p.resolveWg.Wait()

	// Short path so it can get inlined
	if p.isResolved {
		return *p.value
	}

	// Long path
	ch := p.createWaitCh()
	return <-ch
}

func (p *Promise[T]) Then(cb func(T)) {
	// Abuse the goroutine mechanic's to simply await and call the cb in a goroutine
	go func() {
		cb(<-p.Chan())
	}()
}

func (p *Promise[T]) Chan() chan T {
	p.resolveWg.Wait()

	// most common path
	if !p.isResolved {
		return p.createWaitCh()
	}

	// Put this code in a function so the entire chan function most likely can get inlined
	return valueToNewChan(*p.value)
}

func (p *Promise[T]) run(cb func() T) {
	t := cb()

	// If any calls are made to .Then/Await, lock these
	p.resolveWg.Add(1)
	p.isResolved = true
	p.value = &t
	// Note this will be resolved after all waiters are done
	defer p.resolveWg.Done()

	// Ensure no Await/Tech are initializing
	p.awaitWg.Wait()

	p.awaitWg.Add(1)
	defer p.awaitWg.Done()
	for _, ch := range p.chs {
		ch <- t
	}

	// Remove all chan's as it shouldn't be called again
	p.chs = make([]chan T, 0, 1)
}
