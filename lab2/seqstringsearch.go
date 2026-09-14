package main

import (
	"bytes"
	"io"
	"os"
)

// takes in filename, and string to search for, counts how many occurances of given string
func SeqStringSearch(filename string, target string, numChunks int) int {
	file, error := os.Open(filename)
	fileInfo, errorInfo := os.Stat(filename)
	if (error != nil) || (errorInfo != nil) {
		//have to return int, so negative shows error
		return -1
	}

	filelen := fileInfo.Size()
	targetBytes := []byte(target)        //target as byte array
	targetLen := int64(len(targetBytes)) //to get len in byes

	//chunks = # of threads we would have, so change this for
	//comparison w/diff amount of threads in parallel version
	chunks := int64(numChunks)
	chunkSize := filelen / chunks

	totOccurance := 0

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
		chunkBuff := make([]byte, readLen)

		//seek starts reading at given offset and ends, reads chunk
		_, error := file.Seek(startOffset, io.SeekStart)
		if error != nil {
			return -1
		}

		_, error = io.ReadFull(file, chunkBuff)
		if error != nil {
			return -1
		}

		//bytes.Count() iterates through chunk to find # of occurances of target
		chunkOccurance := bytes.Count(chunkBuff, targetBytes)
		totOccurance += chunkOccurance
	}

	file.Close()
	return totOccurance
}
