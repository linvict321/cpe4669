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
		return -1;
	}

	//check if row a = col b
	rowa := len(a)
	colb := len(b[0])

	if rowa != colb{
		fmt.Println("Incompatible matrices")
		return -1;
	}



	//matrix multiplication
	fmt.Println();

}