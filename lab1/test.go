package main

import (
	"fmt"
	"math/rand"
	"time"
)

//generate 2 random matrices
func generateMatrix(rows, cols int) (a, b [][]float32) {
    a = make([][]float32, rows)
    b = make([][]float32, rows)
    
    for i := 0; i < rows; i++ {
        a[i] = make([]float32, cols)
        b[i] = make([]float32, cols)
        for j := 0; j < cols; j++ {
            a[i][j] = rand.Float32()
            b[i][j] = rand.Float32()
        }
    }
    return a, b
}

func compareMatrix (seq, concur [][]float32) bool{
	for i := 0; i < len(seq); i++ {
			for j := 0; j < len(seq[0]); j++ {
				if seq[i][j] != concur[i][j] {
					return false
				}
			}
	}
	return true

}
func RunTests() {
    //generate random matrix once to use for both seq and concur
    A, B := generateMatrix(1000, 1000)
    
    //run seq once
	start := time.Now()
    seqResult := SeqMatrixMult(A, B)
	seqTime := time.Since(start)
	fmt.Printf("Sequential time: %v\n", seqTime)

	//run concur multiple times with different goroutine numbers
	numGo := []int{10, 100, 1000, 10000}
	for i := 0; i < len(numGo); i++{
		numG := numGo[i]
		fmt.Printf("\nTesting with %d goroutines.\n", numG)
		start := time.Now()
		concurResult := ConcurMatrixMult(A, B, numG)
		ConcurMatrixMult(A, B, numG)
		concurTime := time.Since(start)
		fmt.Printf("Concurrent time: %v\n\n", concurTime)
		
		if compareMatrix(seqResult, concurResult) {
			fmt.Printf("Results match.\n")
		} else {
				fmt.Printf("Results are different.\n")
		}
	}
}
