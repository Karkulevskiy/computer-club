package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"

	"github.com/karkulevskiy/computer-club/src/internal"
)

func main() {
	file := MustLoadFile()

	defer file.Close()

	if isValid := internal.ValidateFile(file); !isValid {
		return
	}

	file.Seek(0, io.SeekStart)

	scanner := bufio.NewScanner(file)

	tablesCount, price, start, end, _ := internal.GetOptions(scanner)

	club := internal.NewComputerClub(tablesCount, price, start, end)

	club.StartWork(scanner)
}

func MustLoadFile() *os.File {
	if len(os.Args) != 2 {
		log.Fatalln("test_file is required")
	}

	testFileName := os.Args[1]

	_, err := os.Stat(testFileName)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatalf("test_file %s not found", testFileName)
		}

		log.Fatal(err)
	}

	file, err := os.Open(testFileName)

	if err != nil {
		log.Fatalln(err)
	}

	return file
}
