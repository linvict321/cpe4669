package main

import (
	"bytes"
	"os"
)

// takes in filename, and string to search for, counts how many occurances of given string
func ParStringSearch(filename string, target string, threads int) int {
	file, error := os.Open(filename)
	fileInfo, errorInfo := os.Stat(filename)
	if (error != nil) || (errorInfo != nil) {
		//have to return int, so negative shows error
		return -1
	}

	filelen := fileInfo.Size()
	targetBytes := []byte(target)
	targetLen := int64(len(targetBytes)) //to get len in bytes

	//chunks = # of threads we would have, can change for testing
	chunks := int64(threads)
	chunkSize := filelen / chunks

	totOccurance := 0

	//buff to hold results of all goroutines (count of each chunk)
	resultBuff := make(chan int, chunks)

	for i := int64(0); i < chunks; i++ {
		//since word could be split across two chunks we read
		//the len(target) bytes - 1 extra to account for this
		startOffset := i * chunkSize
		endOffset := startOffset + chunkSize + (targetLen - 1)

		//1st case: last chunk with added targetLen - 1 will exceed filelen
		//2nd case: with how we divided chunks could have remainder bytes not included
		//so make sure last chunk goes to the end of file
		if (endOffset > filelen) || (i == chunks-1) {
			endOffset = filelen
		}

		readLen := endOffset - startOffset

		//goroutines
		go func(start int64, size int64) {
			chunkBuff := make([]byte, size)

			//starts reading at given offset for size of buffer, reads chunk
			_, error := file.ReadAt(chunkBuff, start)
			if error != nil {
				//doesn't count any occurances so returns 0, count will be wrong showing error
				resultBuff <- 0
				return
			}

			//bytes.Count() iterates through chunk to find # of occurances of target
			chunkOccurance := bytes.Count(chunkBuff, targetBytes)
			resultBuff <- chunkOccurance
		}(startOffset, readLen)
	}

	//add up all chunk occurance counts to get total, waits on all goroutines
	for i := 0; i < int(chunks); i++ {
		totOccurance += <-resultBuff
	}

	file.Close()
	return totOccurance
}
