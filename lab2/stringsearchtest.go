package main

import (
	"bytes"
	"fmt"
	"os"
	"time"
)

<<<<<<< HEAD
func test() {
	// file, error := os.Open(filename)
	// if (error != nil) {
	// 	//have to return int, so negative shows error
	// 	return
	// }

=======
func stringSearchTests() {
>>>>>>> 89fd9b67489488476c3cad1070a5a30bec8be72c
	buff, _ := os.ReadFile("test1.txt")

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
	fmt.Printf("Found %d occurances of %s.\n", targetOcc1, target)
	fmt.Printf("Search took %s.\n", elapsed1)

	fmt.Printf("\nRunning parallel string search with %d threads...\n", threads)
	start = time.Now()
	targetOcc2 := ParStringSearch(filename, target, threads)
	elapsed2 := time.Since(start)
	fmt.Printf("Found %d occurances of %s.\n", targetOcc2, target)
	fmt.Printf("Search took %s.\n", elapsed2)

	actualOcc := bytes.Count(buff, []byte("dog"))
	if actualOcc == targetOcc1 && actualOcc == targetOcc2 {
		fmt.Printf("Total occurances of %s is correct.\n", target)
	} else {
		fmt.Printf("Actual occurances of %s is %d, seq and parallel not correct.\n", target, actualOcc)
	}

	//test 2, occurance = 8036
	fmt.Println("\n-------------------Test 2------------------")
	filename = "test1.txt"
	target = "the"
	fmt.Printf("Running sequential string search with %d chunks...\n", chunks)
	start = time.Now()
	targetOcc1 = SeqStringSearch(filename, target, chunks)
	elapsed1 = time.Since(start)
	fmt.Printf("Found %d occurances of %s.\n", targetOcc1, target)
	fmt.Printf("Search took %s.\n", elapsed1)

	fmt.Printf("\nRunning parallel string search with %d threads...\n", threads)
	start = time.Now()
	targetOcc2 = ParStringSearch(filename, target, threads)
	elapsed2 = time.Since(start)
	fmt.Printf("Found %d occurances of %s.\n", targetOcc2, target)
	fmt.Printf("Search took %s.\n", elapsed2)

	actualOcc = bytes.Count(buff, []byte("the"))
	if actualOcc == targetOcc1 && actualOcc == targetOcc2 {
		fmt.Printf("Total occurances of %s is correct.\n", target)
	} else {
		fmt.Printf("Actual occurances of %s is %d, seq and parallel not correct.\n", target, actualOcc)
	}

	//test 3, occurance = 12053
	fmt.Println("\n-------------------Test 3------------------")
	filename = "test1.txt"
	target = "my"
	fmt.Printf("Running sequential string search with %d chunks...\n", chunks)
	start = time.Now()
	targetOcc1 = SeqStringSearch(filename, target, chunks)
	elapsed1 = time.Since(start)
	fmt.Printf("Found %d occurances of %s.\n", targetOcc1, target)
	fmt.Printf("Search took %s.\n", elapsed1)

	fmt.Printf("\nRunning parallel string search with %d threads...\n", threads)
	start = time.Now()
	targetOcc2 = ParStringSearch(filename, target, threads)
	elapsed2 = time.Since(start)
	fmt.Printf("Found %d occurances of %s.\n", targetOcc2, target)
	fmt.Printf("Search took %s.\n", elapsed2)

	actualOcc = bytes.Count(buff, []byte("my"))
	if actualOcc == targetOcc1 && actualOcc == targetOcc2 {
		fmt.Printf("Total occurances of %s is correct.\n", target)
	} else {
		fmt.Printf("Actual occurances of %s is %d, seq and parallel not correct.\n", target, actualOcc)
	}
}
