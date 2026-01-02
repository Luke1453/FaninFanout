package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type DoneChan <-chan struct{}

func repeatFunc[T any](done DoneChan, fn func() T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case out <- fn():
			}
		}
	}()

	return out
}

func take[T any](done DoneChan, in <-chan T, n int) <-chan T {
	taken := make(chan T)
	go func() {
		defer close(taken)
		for i := 0; i < n; i++ {
			select {
			case <-done:
				return
			case taken <- <-in:
			}
		}
	}()

	return taken
}

func fanIn[T any](done DoneChan, chans ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	fannedInStream := make(chan T)

	transfer := func(in <-chan T) {
		defer wg.Done()
		for i := range in {
			select {
			case <-done:
				return
			case fannedInStream <- i:
			}
		}
	}

	for _, in := range chans {
		wg.Add(1)
		go transfer(in)
	}

	go func() {
		wg.Wait()
		close(fannedInStream)
	}()

	return fannedInStream
}

func isPrime(n int) bool {
	for i := 2; i < n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func primeFinder(done DoneChan, randIntStream <-chan int) <-chan int {
	primes := make(chan int)
	go func() {
		defer close(primes)
		for {
			select {
			case <-done:
				return
			case randInt := <-randIntStream:
				if isPrime(randInt) {
					primes <- randInt
				}
			}
		}
	}()

	return primes
}

func main() {
	start := time.Now()

	done := make(chan struct{})
	defer close(done)

	randomNumberGenerator := func() int { return rand.Intn(1_000_000_000) }
	randomStream := repeatFunc(done, randomNumberGenerator)

	// fan out
	CpuCount := runtime.NumCPU()
	primeFinderChannels := make([]<-chan int, CpuCount)
	for i := 0; i < CpuCount; i++ {
		primeFinderChannels[i] = primeFinder(done, randomStream)

	}

	// fan in
	fannedInPrimeStream := fanIn(done, primeFinderChannels...)
	for rando := range take(done, fannedInPrimeStream, 100) {
		fmt.Println(rando)
	}

	fmt.Println(time.Since(start))
}
