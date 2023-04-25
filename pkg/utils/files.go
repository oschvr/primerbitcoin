package utils

import (
	"fmt"
	"log"
	"os"
)

// ListDirFiles list files in a directory
func ListDirFiles(path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Println(f.Name())
	}
}
