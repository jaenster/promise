package promise

func valueToNewChan[T interface{}](v T) (ch chan T) {
	ch = make(chan T)
	go func() {
		ch <- v
	}()
	return ch
}
