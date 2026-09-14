package main

import (
	"fmt"
	"time"
)

const (
	filename = "test1.txt"
	runs = 10
)

//times how long each run takes for a function
func timing(function func()) time.Duration {
	start := time.Now()
	function()
	return time.Since(start)
}

//averages the time for running a function
func avgDuration(function func()) time.Duration {
	total := time.Duration(0)

	for i := 0; i < runs; i++ {
		total += timing(function)
	}

	return total / time.Duration(runs)
}

func main() {
	words := []string{"dog", "the", "my"}
	threadCounts := []int{1, 2, 4, 8, 16, 32, 100}


	//test each word
	for i := 0; i < len(words); i++ {
		word := words[i]

		fmt.Printf("\n=== word: %q ===\n", word)

		//run and time sequential search
		seqCount := SeqStringSearch(filename, word, 1)
		seqTime := avgDuration(func() {
			SeqStringSearch(filename, word, 1)
		})

		fmt.Printf("threads | seq        | par         | speedup  | count\n")

		fmt.Printf("seq     | %-10v | —           | -        | %d\n",
			seqTime, seqCount)

		//test parallel search with different number of threads
		for i := 0; i < len(threadCounts); i++ {
			n := threadCounts[i]

			//time parallel search
			parCount := ParStringSearch(filename, word, n)
			parTime := avgDuration(func() {
				ParStringSearch(filename, word, n)
			})

			speedup := float64(seqTime) / float64(parTime)

			if parCount != seqCount {
				fmt.Printf("%-7d | —          | %-11v | %.2fx    | %d MISMATCH\n",
					n, parTime, speedup, parCount)
			} else {
				fmt.Printf("%-7d | —          | %-11v | %.2fx    | %d\n",
					n, parTime, speedup, parCount)
			}
		}
	}
}
