package main

import (
	"fmt"
	"math"
)

func main(){

	//example matrices
	a := [][]uint8{
		{0, 1, 2, 3},
		{4, 5, 6, 7}
	}

	b := [][]uint8{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{9, 10, 11}
	}

	//check if matrix a & b are valid
	if len(a) == 0 || len(b) == 0 || len(a[0]) == 0 || len(b[0]){
		fmt.Println("Invalid matrices");
		return -1
	}

	//check if len of row a =  len of col b
	rowa := len(a)
	colb := len(b[0])

	if rowa != colb{
		fmt.Println("Incompatible matrices")
		return -1
	}

	//initialize result (cola x rowb), start filling matrix
	result := [len(a[0])][len(b)]{}

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
	fmt.Println()

}