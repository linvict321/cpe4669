//sequential version
package main

import (
	"fmt"
)

func SeqMatrixMult(a [][]int, b [][]int) ([][]int) {
	//check if matrix a & b are valid
	if len(a) == 0 || len(b) == 0 || len(a[0]) == 0 || len(b[0]) == 0 {
		fmt.Println("Invalid matrices");
		return nil
	}

	//check if len of row a =  ken of col b
	rowa := len(a[0])
	colb := len(b)

	if rowa != colb{
		fmt.Println("Incompatible matrices")
		return nil
	}

	//initialize result (rowa x colb), start filling matrix
	result := make([][]int, rowa)
	for i := range result {
		result[i] = make([]int, colb)
	}

	//loop through row a, multiply it into row b
	for i := 0; i < len(a); i++{
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

func main(){

	//example matrices
	a := [][]int{
		{10, 1, 2, 3},
		{4, 5, 6, 7},
	}

	b := [][]int{
		{10, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{9, 10, 11},
	}

	result := SeqMatrixMult(a, b)

	for i:= range result{
		fmt.Println(result[i])
	}

}

