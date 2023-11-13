package main

import "fmt"

const N = 20

func main() {
	done := make(chan int, 1)

	for i := 0; i < N; i++ {
		go func(done chan int) {
			a := 0
			b := 0
			for {
				b++
				a += b
				if b > 1_000_000_000/N {
					break
				}
			}

			done <- a
		}(done)
	}

	for i := 0; i < N; i++ {
		fmt.Printf("%d\n", <-done)
	}
}
