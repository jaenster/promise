# Promises

Go has great concurrency, but (javascript like) promises some time compliment go routines. Go routines are great, but it
has no ability to wait on the result.

While its possible to work with channels, sometimes it can be more clean (code wise) to write it in a promise form.

Specially if a big list of goroutines run, a simple promise.Race or promise.All can be just what you wanted.

## Why not just use a channel?

The nice thing of a promise is that a promise can already be fulfilled. Running an `.Await` on a promise that is already
resolved, just instantly gives back the resolved value. While a channel will wait again for the next message (most likely indefinitely with a single use response)

## Callbacks

While callbacks are used in javascript for promises, it doesn't make much sense to use actual callback behaviour in go,
as blocking code in a goroutine is a perfectly valid use case. Therefor the decision has been made to simply use the
return value of the new promise as return value

## Await / then

Since callbacks don't make the most sense in go, `.Then` is one of the functions on a promise. While in go the use
of `.Await` seems to be more useful

## Reject? Error handling?

While error handling is nice, its recommend using pointers for models and return values. A simple `nil` in the response
over some error that makes things very complex and unreliable and slower. If the database goes down while fetching a
model, a panic feels right. When a model isn't found; returning `nil` feels accurate.

## Lib supports

### Creating promises

- `promise.New[T any](cb func() T) *Promise[T]`
    - Will start cb as goroutine
- `promise.FromChannel[T any](ch <-chan T) *Promise[T]`
    - Simple wrapper around New that resolves once the channel gives a result
- `promise.All[T any](ps ...*Promise[T]) *Promise[[]T]`
    - Return of *promise[T[]], accepts promises of the same type
- `promise.Race[T any](ps ...*Promise[T]) *Promise[T]`
    - Returns a promise that resolves with the first promise of the array that resolves
- `Resolve[T any](value T) *Promise[T]`
    - Returns an already resolved promise

### Methods on promises

- `p.Await() T`
    - sync function that returns the value of the resolved promise
- `p.Then(func(v T))`
    - supply a callback that gets called once the promise is resolved
- `p.Chan() chan T`
    - returns a channel that supplies the data once the promise is resolved.
    - Note; if promise already resolved it makes a single use channel for it

## Example

```go
package main

import (
  "fmt"
  "github.com/jaenster/promise"
  "time"
)

type Model struct {
  Id uint64
  X  uint32
  Y  uint32
}

func getModelFromDatabase(id uint64) *promise.Promise[*Model] {
  return promise.New(func() *Model {

    // ... Some code that suppose to get it
    // Blocking code is perfectly valid in go, unlike javascript
    time.Sleep(50 * time.Millisecond)

    // Mock id 2 not being found
    if id == 2 {
      return nil
    }

    // Return instead of resolve
    model := Model{id, 0, 0}
    return &model
  })
}

func main() {
  // Example #1
  // Silly example, as go is fine with blocking code, this could be done without promises just fine
  value := getModelFromDatabase(1).Await()
  fmt.Sprintln(value)

  // Example #2
  // This gets a bit more complex to write with multiple channels and query in native go

  // Start different queries for models
  p := promise.All(getModelFromDatabase(1), getModelFromDatabase(2))

  // Await all results which returns the correctly sorted results as put in at All
  results := p.Await()

  one := results[0]
  two := results[1]

  fmt.Sprintln(one)
  fmt.Sprintln(two)
}

```