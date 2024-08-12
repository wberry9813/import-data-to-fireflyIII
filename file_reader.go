package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	// ReadCSVFiles reads all CSV files from the specified directory.
	"sort"
)

func ReadCSVFiles(directory string) []string {
	files, err := func() ([]fs.FileInfo, error) {
		f, err := os.Open(directory)
		if err != nil {
			return nil, err
		}
		list, err := f.Readdir(-1)
		f.Close()
		if err != nil {
			return nil, err
		}
		sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })
		return list, nil
	}()
	if err != nil {
		log.Fatalf("Failed to read directory %s: %v", directory, err)
	}

	var csvFiles []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".csv" {
			csvFiles = append(csvFiles, filepath.Join(directory, file.Name()))
		}
	}

	return csvFiles
}
