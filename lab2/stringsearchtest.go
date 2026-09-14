package main

import (
	"fmt"
	"time"
)

func main() {
	//test 1, occurance = 4018
	filename := "test1.txt"
	target := "dog"
	fmt.Println("Running sequential string search...")
	start := time.Now()
	targetOcc1 := SeqStringSearch(filename, target)
	elapsed1 := time.Since(start)
	fmt.Printf("\tSearched file %s.\nFound %d occurances of %s.\n", filename, targetOcc1, target)
	fmt.Printf("\tSearch took %s.\n", elapsed1)

	threads := 10 //**Change for testing diff number of threads, 1, 10, 50, 100 maybe**
	fmt.Printf("Running parallel string search with %d threads...\n", threads)
	start = time.Now()
	targetOcc2 := ParStringSearch(filename, target, threads)
	elapsed2 := time.Since(start)
	fmt.Printf("\tSearched file: %s with %d threads.\nFound %d occurances of %s.\n", filename, threads, targetOcc2, target)
	fmt.Printf("\tSearch took %s.\n", elapsed2)

	if targetOcc1 == targetOcc2 {
		fmt.Println("Parallel and sequential string search occurances matched!")
	} else {
		fmt.Println("Parallel and sequential string search occurances did not match.")
	}

	//test 2, occurance = 8036
	filename = "test1.txt"
	target = "the"
	fmt.Println("\nRunning sequential string search...")
	start = time.Now()
	targetOcc1 = SeqStringSearch(filename, target)
	elapsed1 = time.Since(start)
	fmt.Printf("\tSearched file %s.\nFound %d occurances of %s.\n", filename, targetOcc1, target)
	fmt.Printf("\tSearch took %s.\n", elapsed1)

	threads = 10 //**Change for testing diff number of threads, 1, 10, 50, 100 maybe**
	fmt.Printf("Running parallel string search with %d threads...\n", threads)
	start = time.Now()
	targetOcc2 = ParStringSearch(filename, target, threads)
	elapsed2 = time.Since(start)
	fmt.Printf("\tSearched file: %s with %d threads.\nFound %d occurances of %s.\n", filename, threads, targetOcc2, target)
	fmt.Printf("\tSearch took %s.\n", elapsed2)

	if targetOcc1 == targetOcc2 {
		fmt.Println("Parallel and sequential string search occurances matched!")
	} else {
		fmt.Println("Parallel and sequential string search occurances did not match.")
	}

	//test 3, occurance = 12053
	filename = "test1.txt"
	target = "my"
	fmt.Println("\nRunning sequential string search...")
	start = time.Now()
	targetOcc1 = SeqStringSearch(filename, target)
	elapsed1 = time.Since(start)
	fmt.Printf("\tSearched file %s.\nFound %d occurances of %s.\n", filename, targetOcc1, target)
	fmt.Printf("\tSearch took %s.\n", elapsed1)

	threads = 10 //**Change for testing diff number of threads, 1, 10, 50, 100 maybe**
	fmt.Printf("Running parallel string search with %d threads...\n", threads)
	start = time.Now()
	targetOcc2 = ParStringSearch(filename, target, threads)
	elapsed2 = time.Since(start)
	fmt.Printf("\tSearched file: %s with %d threads.\nFound %d occurances of %s.\n", filename, threads, targetOcc2, target)
	fmt.Printf("\tSearch took %s.\n", elapsed2)

	if targetOcc1 == targetOcc2 {
		fmt.Println("Parallel and sequential string search occurances matched!")
	} else {
		fmt.Println("Parallel and sequential string search occurances did not match.")
	}
}
