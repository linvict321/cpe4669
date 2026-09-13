package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
)

// takes in filename, and string to search for, counts how many occurances of given string
func SeqStringSearch(filename string, target string) int {
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
	chunks := int64(100)
	chunkSize := filelen / chunks

	totOccurance := 0

	//int j = 0;
	//for i to chunks
	//	int j = j * i; //gives start index of chunk
	//	for j to j + chunkSize
	//		count how many occurances of word
	//		after for loop add chunk size to j, so next j starts at start of next chunk

	for i := int64(0); i < chunks; i++ {
		//since word could be split across two chunks we read
		//the len(target) bytes - 1 extra to account for this
		startOffset := i * chunkSize
		endOffset := startOffset + chunkSize + (targetLen - 1)

		if endOffset > filelen {
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

func main() {
	filename := "test1.txt"
	target := "dog"
	start := time.Now()
	targetOcc := SeqStringSearch(filename, target)
	elapsed := time.Since(start)
	fmt.Printf("Searched file: %s\nFound %d occurances of %s.\n", filename, targetOcc, target)
	fmt.Printf("Search took %s.\n", elapsed)
}
