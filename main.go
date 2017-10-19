package main

import (
	"flag"
	"fmt"
	"time"
)

func max(xs []int) int {
	if len(xs) < 1 {
		panic("max() on empty slice")
	}
	ret := xs[0]
	for i := 1; i < len(xs); i++ {
		if xs[i] > ret {
			ret = xs[i]
		}
	}
	return ret
}

func main() {
	n := flag.Int("nodes", 5, "number of participating nodes")
	iters := flag.Int("iters", 100000, "number of iterations")
	debug := flag.Bool("debug", false, "print debug trace")

	flag.Parse()

	// 1. process i is in the doorway while choosing[i] == true
	// 2. process i is in the bakery from when it resets choosing[i] to false till
	// either it fails, or finishes its critical section.

	choosing := make([]bool, *n+1, *n+1) // one extra for the monitoring process
	numbers := make([]int, *n+1, *n+1)

	sharedMap := map[string]int{
		"last_updated_by": -1,
	}

	sharedSlice := make([]int, 0)

	lock := func(id int) {

		choosing[id] = true

		/* At the doorway */

		if *debug {
			fmt.Printf("%d: choosing\n", id)
		}

		numbers[id] = 1 + max(numbers)
		if *debug {
			fmt.Printf("%d: chose number %d\n", id, numbers[id])
		}

		choosing[id] = false

		/* Entered bakery */

		for i := 0; i <= *n; i++ {
			printed := false
			for choosing[i] {
				// spin
				if *debug && !printed {
					fmt.Printf("%d: spinning for %d (waiting for it to choose: %v)\n", id, i, choosing)
					printed = true
				}
				// this is important, as tight spinning here would cause only
				// the first $n_core goroutines to be scheduled, starving
				// all others. this would in turn cause the whole algorithm
				// to stop making any progress.
				time.Sleep(10 * time.Millisecond)
			}
			printed = false
			for numbers[i] != 0 && ((numbers[i] < numbers[id]) || (numbers[i] == numbers[id] && i < id)) {
				// spin
				if *debug && !printed {
					fmt.Printf("%d: spinning for %d (waiting for it to finish CS: %v)\n", id, i, numbers)
					printed = true
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}

	unlock := func(id int) {
		if *debug {
			fmt.Printf("%d: unlock: %v\n", id, numbers)
		}
		numbers[id] = 0
	}

	proc := func(id int, stop chan struct{}, ack chan struct{}) {
	out:
		for {
			select {
			case <-stop:
				break out

			default:

				lock(id)
				fmt.Printf("%d: entering critical section\n", id)

				// critical section
				sharedMap["last_updated_by"] = id
				sharedSlice = append(sharedSlice, id)
				sharedSlice = sharedSlice[1:]
				// end critical section

				fmt.Printf("%d: done critical section\n", id)
				unlock(id)
			}
		}
		ack <- struct{}{}
	}

	stops := make([]chan struct{}, 0)
	acks := make([]chan struct{}, 0)
	for i := 0; i < *n; i++ {
		c := make(chan struct{})
		d := make(chan struct{})
		stops = append(stops, c)
		acks = append(acks, d)
		go proc(i, c, d)
	}

	// Monitor and quit after some time

	for i := 1; i < *iters; i++ {
		lock(*n)
		if *debug {
			fmt.Sprintf("monitor lock acquired", sharedSlice)
		}
		if len(sharedSlice) > 1 {
			panic(fmt.Sprintf("found concurrent access with %v", sharedSlice))
		}
		// Go will panic when sharedMap is written to by two goroutines

		unlock(*n)
		if *debug {
			fmt.Sprintf("monitor lock released", sharedSlice)
		}
	}

	for _, stop := range stops {
		fmt.Println("stopping")
		stop <- struct{}{}
	}

	for _, ack := range acks {
		<-ack
	}

}
