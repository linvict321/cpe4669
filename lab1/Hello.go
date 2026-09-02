// sequential version
package main

import (
	"fmt"
)

func SeqMatrixMult(a [][]float32, b [][]float32) ([][]float32) {
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

	//initialize result (rowa x colb), start filling matrix
	result := make([][]float32, rowsa)
	for i := range result {
		result[i] = make([]float32, colsb)
	}

	//loop through row a, multiply it into row b
	for i := 0; i < len(a); i++{
		if i % 10 == 0 {
			fmt.Printf("Row: %d\r", i)
		}
		for j := 0; j < len(b[0]); j++{
			//start filling by 
			for k := 0; k < len(b); k++{
				result[i][j] += a[i][k] * b[k][j]
			}
		}
	}

	//matrix multiplication
	return result

}

func main() {
	RunTests() // main from tests.go

	// //example matrices
	// a := [][]float32{
	// 	{10, 1, 2, 3},
	// 	{4, 5, 6, 7},
	// }

	// b := [][]float32{
	// 	{10, 1, 2},
	// 	{3, 4, 5},
	// 	{6, 7, 8},
	// 	{9, 10, 11},
	// }

	// result := SeqMatrixMult(a, b)
	
	// fmt.Println("-----SEQUENTIAL-----")
	// fmt.Println("\nResult 1 - a*b:")
	// for i:= range result{
	// 	fmt.Println(result[i])
	// }
	
	// c := [][]float32 {
	// 	{1},
	// 	{5},
	// }

	// d := [][]float32 {
	// 	{1, 2},
	// }

	// result = SeqMatrixMult(c, d)
	
	// fmt.Println("\nResult 2 - c*d:")
	// for i:= range result{
	// 	fmt.Println(result[i])
	// }

	// fmt.Println("\n-----CONCURRENT-----")

	// result = ConcurMatrixMult(a, b, 10000)
	
	// fmt.Println("\nResult 1 - a*b:")
	// for i:= range result{
	// 	fmt.Println(result[i])
	// }


	// result = ConcurMatrixMult(c, d, 10000)
	
	// fmt.Println("\nResult 2 - c*d:")
	// for i:= range result{
	// 	fmt.Println(result[i])
	// }

}

