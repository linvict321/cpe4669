package main

import (
	"fmt"
	"time"
)

func main() {
	//**Change for testing diff number of threads, 1, 10, 50, 100 maybe**
	threads := 10
	chunks := 10

	//test 1, occurance = 4018
	fmt.Println("\n-------------------Test 1------------------")
	filename := "test1.txt"
	target := "dog"
	fmt.Printf("Running sequential string search with %d chunks...\n", chunks)
	start := time.Now()
	targetOcc1 := SeqStringSearch(filename, target, chunks)
	elapsed1 := time.Since(start)
	fmt.Printf("Searched file %s.\nFound %d occurances of %s.\n", filename, targetOcc1, target)
	fmt.Printf("Search took %s.\n", elapsed1)

	fmt.Printf("\nRunning parallel string search with %d threads...\n", threads)
	start = time.Now()
	targetOcc2 := ParStringSearch(filename, target, threads)
	elapsed2 := time.Since(start)
	fmt.Printf("Searched file: %s with %d threads.\nFound %d occurances of %s.\n", filename, threads, targetOcc2, target)
	fmt.Printf("Search took %s.\n", elapsed2)

	if targetOcc1 == targetOcc2 {
		fmt.Println("Parallel and sequential string search occurances matched!")
	} else {
		fmt.Println("Parallel and sequential string search occurances did not match.")
	}

	//test 2, occurance = 8036
	fmt.Println("\n-------------------Test 2------------------")
	filename = "test1.txt"
	target = "the"
	fmt.Printf("Running sequential string search with %d chunks...\n", chunks)
	start = time.Now()
	targetOcc1 = SeqStringSearch(filename, target, chunks)
	elapsed1 = time.Since(start)
	fmt.Printf("Searched file %s.\nFound %d occurances of %s.\n", filename, targetOcc1, target)
	fmt.Printf("Search took %s.\n", elapsed1)

	fmt.Printf("\nRunning parallel string search with %d threads...\n", threads)
	start = time.Now()
	targetOcc2 = ParStringSearch(filename, target, threads)
	elapsed2 = time.Since(start)
	fmt.Printf("Searched file: %s with %d threads.\nFound %d occurances of %s.\n", filename, threads, targetOcc2, target)
	fmt.Printf("Search took %s.\n", elapsed2)

	if targetOcc1 == targetOcc2 {
		fmt.Println("Parallel and sequential string search occurances matched!")
	} else {
		fmt.Println("Parallel and sequential string search occurances did not match.")
	}

	//test 3, occurance = 12053
	fmt.Println("\n-------------------Test 3------------------")
	filename = "test1.txt"
	target = "my"
	fmt.Printf("Running sequential string search with %d chunks...\n", chunks)
	start = time.Now()
	targetOcc1 = SeqStringSearch(filename, target, chunks)
	elapsed1 = time.Since(start)
	fmt.Printf("Searched file %s.\nFound %d occurances of %s.\n", filename, targetOcc1, target)
	fmt.Printf("Search took %s.\n", elapsed1)

	fmt.Printf("\nRunning parallel string search with %d threads...\n", threads)
	start = time.Now()
	targetOcc2 = ParStringSearch(filename, target, threads)
	elapsed2 = time.Since(start)
	fmt.Printf("Searched file: %s with %d threads.\nFound %d occurances of %s.\n", filename, threads, targetOcc2, target)
	fmt.Printf("Search took %s.\n", elapsed2)

	if targetOcc1 == targetOcc2 {
		fmt.Println("Parallel and sequential string search occurances matched!")
	} else {
		fmt.Println("Parallel and sequential string search occurances did not match.")
	}
}
