package promise

import (
	"runtime"
	"testing"
	"time"
)

type empty struct{}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func Ptr[T any](v T) *T {
	return &v
}

func TestPack(t *testing.T) {
	type myStruct struct {
		X int64
		Y int64
	}

	t.Run("Promise await", func(t *testing.T) {
		p := New(func() int {
			time.Sleep(50 * time.Millisecond)
			return 1337
		})

		before := makeTimestamp()
		v := p.Await()
		diff := makeTimestamp() - before

		if v != 1337 {
			t.Error("Value corrupt")
			t.Fail()
		}

		if diff < 50 {
			t.Error("Promise takes less as 50 ms to run")
			t.Fail()
		}
	})

	t.Run("Promise channel", func(t *testing.T) {
		p := New(func() int {
			time.Sleep(50 * time.Millisecond)
			return 1337
		})

		before := makeTimestamp()
		v := <-p.Chan()
		diff := makeTimestamp() - before

		if v != 1337 {
			t.Error("Value corrupt")
			t.Fail()
		}

		if diff < 50 {
			t.Error("Promise takes less as 50 ms to run")
			t.Fail()
		}
	})

	t.Run("Promise then", func(t *testing.T) {
		p := New(func() int {
			time.Sleep(50 * time.Millisecond)
			return 1337
		})

		data := make(chan int)
		p.Then(func(v int) {
			data <- v
		})

		before := makeTimestamp()
		v := <-data
		diff := makeTimestamp() - before

		if v != 1337 {
			t.Error("Value corrupt")
			t.Fail()
		}

		if diff < 50 {
			t.Error("Promise takes less as 50 ms to run")
			t.Fail()
		}
	})

	t.Run("Promise", func(t *testing.T) {
		p := New(func() int {
			time.Sleep(50 * time.Millisecond)
			return 1337
		})

		before := makeTimestamp()
		v := p.Await()
		diff := makeTimestamp() - before

		if v != 1337 {
			t.Error("Value corrupt")
			t.Fail()
		}

		if diff < 50 {
			t.Error("Promise takes less as 50 ms to run")
			t.Fail()
		}
	})

	t.Run("typical", func(t *testing.T) {
		resolve := make(chan empty)
		waitThen := make(chan empty)

		p := New(func() int {
			<-resolve
			return 1337
		})

		p.Then(func(v int) {
			waitThen <- empty{}
		})

		if p.isResolved == true {
			t.Error("Promise is resolved when it should not be")
			t.Fail()
		}

		// signal promise it is ready to resolve
		resolve <- empty{}

		// Wait until it is resolved
		<-waitThen
		if p.isResolved == false {
			t.Error("Promise is not resolved when it should be")
		}

		if *p.value != 1337 {
			t.Error("Promise value is corrupt")
		}

		if len(p.chs) > 1 {
			t.Error("More as zero channels still there")
		}
	})

	t.Run("After garbage collector ran", func(t *testing.T) {
		resolve := make(chan empty)
		waitThen := make(chan empty)

		p := New(func() *myStruct {
			<-resolve
			return &myStruct{5, 5}
		})

		p.Then(func(v *myStruct) {
			waitThen <- empty{}
		})

		if p.isResolved == true {
			t.Error("Promise is resolved when it should not be")
			t.Fail()
		}

		// signal promise it is ready to resolve
		resolve <- empty{}

		// Wait until it is resolved
		<-waitThen

		// Ensure its gone
		for i := 0; i < 10; i++ {
			runtime.GC()
		}

		if p.isResolved == false {
			t.Error("Promise is not resolved when it should be")
		}

		if len(p.chs) > 1 {
			t.Error("More as zero channels still there")
		}
	})

	t.Run("All", func(t *testing.T) {
		results := All(
			New(func() int {
				time.Sleep(100 * time.Millisecond)
				return 65
			}),
			New(func() int {
				return 75
			}),
		).Await()

		a := results[0]
		if a != 65 {
			t.Error("Promise one doesnt give back the proper response")
		}
		b := results[1]
		if b != 75 {
			t.Error("Promise two doesnt give back the proper response")
		}

	})

	t.Run("Race", func(t *testing.T) {
		promises := make([]*Promise[int], 2, 2)
		one := New(func() int {
			time.Sleep(100 * time.Millisecond)
			return 65
		})
		two := New(func() int {
			return 75
		})
		promises[0] = one
		promises[1] = two

		first := Race(promises).Await()
		if first != 75 {
			t.Error("Slowest promise won")
		}
	})
}
