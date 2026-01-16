---
title: Understanding Concurrency in Go
date: 2026-01-10T09:15:00Z
description: Learn how Go's concurrency primitives - goroutines and channels - make concurrent programming simple and elegant
tags:
  - go
  - concurrency
  - tutorial
  - advanced
---

# Understanding Concurrency in Go

One of Go's most powerful features is its built-in support for concurrent programming through goroutines and channels.

## Goroutines: Lightweight Threads

A goroutine is a lightweight thread managed by the Go runtime. Starting a goroutine is as simple as adding the `go` keyword:

```go
func main() {
    go doSomething()  // Runs concurrently
    doSomethingElse() // Runs in main goroutine
}
```

## Channels: Communication Between Goroutines

Channels provide a way for goroutines to communicate with each other:

```go
func main() {
    messages := make(chan string)

    go func() {
        messages <- "ping"
    }()

    msg := <-messages
    fmt.Println(msg)
}
```

## Buffered Channels

Channels can be buffered to allow sending multiple values without blocking:

```go
ch := make(chan int, 2)
ch <- 1
ch <- 2
// These don't block because buffer has capacity
```

## Select Statement

The `select` statement lets you wait on multiple channel operations:

```go
select {
case msg1 := <-ch1:
    fmt.Println("Received", msg1)
case msg2 := <-ch2:
    fmt.Println("Received", msg2)
case <-time.After(time.Second):
    fmt.Println("Timeout")
}
```

## Common Patterns

### Worker Pool

```go
func worker(id int, jobs <-chan int, results chan<- int) {
    for j := range jobs {
        results <- j * 2
    }
}

func main() {
    jobs := make(chan int, 100)
    results := make(chan int, 100)

    // Start 3 workers
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }

    // Send jobs
    for j := 1; j <= 5; j++ {
        jobs <- j
    }
    close(jobs)

    // Collect results
    for a := 1; a <= 5; a++ {
        <-results
    }
}
```

## Best Practices

1. Always close channels when you're done sending
2. Use `sync.WaitGroup` for synchronization
3. Be careful of deadlocks - ensure goroutines can complete
4. Use `context` for cancellation and timeouts

Concurrency in Go is powerful but requires careful thought about data flow and synchronization!
