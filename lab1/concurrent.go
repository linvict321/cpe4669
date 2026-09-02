package main

import (
	"fmt"
)

type CellResult struct {
	result float64
	row int
	col int
}

// helper function, which returns the matrix result from multiplying
// a row of A by just one col of B.
// feeds a matrix the same size as the overall solution into a channel
func calcCellResultAndInputToChannel (a [][]float64, b [][]float64, rowa int, colb int, ch chan CellResult) {
	//initialize result (rowa x colb), start filling matrix
	cr := CellResult{
		result: 0,
		row: rowa,
		col: colb,
	}

	// elt-wise mult: one row of a, one col of b
	for i := range len(b) {
		cr.result += a[rowa][i] * b[i][colb]
	}
	ch <- cr
}

func calcFinalMatFromCellResults (ch chan CellResult, rows int, cols int) [][]float64 {
	// initialize final mat with zeros
	final := make([][]float64, rows)
	for i := range rows {
		final[i] = make([]float64, cols)
	}

	for range rows*cols {
		cr := <- ch
		final[cr.row][cr.col] += cr.result
	}

	return final
}

func ConcurMatrixMult(a [][]float64, b [][]float64, numGoroutines int) [][]float64 {
	//check if matrix a & b are valid
	if len(a) == 0 || len(b) == 0 || len(a[0]) == 0 || len(b[0]) == 0 {
		fmt.Println("Invalid matrices");
		return nil
	}

	//check if len of row a = len of col b (#cols a == #rows b)
	rowsa := len(a)
	colsa := len(a[0])
	rowsb := len(b)
	colsb := len(b[0])

	if colsa != rowsb{
		fmt.Println("Incompatible matrices")
		return nil
	}

	ch := make(chan CellResult, rowsa*colsb)
	goroutineLimit := make(chan struct{}, numGoroutines)

	for i := range rowsa {
		if i % 10 == 0 {
			fmt.Printf("Row: %d\r", i)
		}
		for j := range colsb {
			goroutineLimit <- struct{}{} // blocks if max num goroutines are running
			
			// wrap in this function to adhere to the goRoutine limit 
			go func(i, j int) {
				defer func() { <-goroutineLimit }() // defer requires a function to work
				
				calcCellResultAndInputToChannel(a, b, i, j, ch)
			}(i, j)
		}
	}

	final := calcFinalMatFromCellResults(ch, rowsa, colsb)
	return final
}