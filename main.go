package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

func main() {
	var fIn, fOut, tCountStr string
	var tCount int
	var checkedList []string

	flag.StringVar(&fIn, "i", "", "")
	flag.StringVar(&fOut, "o", "", "")
	flag.StringVar(&tCountStr, "t", "", "")
	flag.Parse()

	tCount, err := strconv.Atoi(tCountStr)
	if err != nil {
		fmt.Println("error: wrong cli param")
		os.Exit(1)
	}

	if !exists(fIn) {
		fmt.Println("error: read input file")
		os.Exit(1)
	}

	if !exists(fOut) {
		err := TouchFile(fOut)
		if err != nil {
			fmt.Println("error: create out file")
			os.Exit(1)
		}
	}

	if (tCount == 0) || (tCount > 1024) {
		fmt.Println("error: wrong thread config")
		os.Exit(1)
	}

	lines, err := readLines(fIn)
	if (err != nil) || (len(lines) == 0) {
		fmt.Println("error: empty input file")
		os.Exit(1)
	}

	fmt.Println("Begin Check")
	checkedList = doSyncCheck(lines, tCount)

	if len(checkedList) > 0 {
		f, err := os.OpenFile(fOut, os.O_RDWR, 0644)
		if err == nil {
			for _, Line := range checkedList {
				f.WriteString(Line + "\n")
			}
			f.Close()
		} else {
			fmt.Println("error: write output file", err)
			os.Exit(1)
		}
	}
}
