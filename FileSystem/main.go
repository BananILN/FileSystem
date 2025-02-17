package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func getFilesAndSizes(root string) ([]string, []int64, error) {
	var files []string
	var sizes []int64

	fmt.Printf("Scanning directory: %s\n", root)

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем корневую директорию
		if path == root {
			return nil
		}

		// Проверяем, является ли файл или директория на первом уровне
		if filepath.Dir(path) == root {
			if info.IsDir() {
				size := getDirSize(path) // Изменено на получение одного значения
				files = append(files, path)
				sizes = append(sizes, size)
			} else {
				files = append(files, path)
				sizes = append(sizes, info.Size())
			}
		}

		return nil
	})

	return files, sizes, err
}

func getDirSize(path string) int64 {
	var size int64

	// Рекурсивно обходим все файлы и поддиректории.
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Для файлов добавляем их размер.
			if info.Name() != filepath.Base(path){
				size += info.Size()
			}
		} else{
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		fmt.Println("ошибка при вычислении размера директории:", err)
		return 0
	}

	return size
}

func sortFiles(files []string, sizes []int64, order string) {
	type fileSize struct {
		path string
		size int64
	}
	var fileSizes []fileSize
	for i := range files {
		fileSizes = append(fileSizes, fileSize{files[i], sizes[i]})
	}

	sort.Slice(fileSizes, func(i, j int) bool {
		if order == "asc" {
			return fileSizes[i].size < fileSizes[j].size
		} else {
			return fileSizes[i].size > fileSizes[j].size
		}
	})

	for i := 0; i < len(fileSizes); i++ {
		files[i] = fileSizes[i].path
		sizes[i] = fileSizes[i].size
	}
}

func formatSize(size int64) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

func printFiles(files []string, sizes []int64) error {
	if len(files) == 0 {
		fmt.Println("No files or directories found.")
		return nil
	}

	for i, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			fmt.Printf("Error getting info for %s: %s\n", file, err)
			return err
		}

		name := filepath.Base(file)
		sizeFormatted := formatSize(sizes[i])

		if info.IsDir() {
			fmt.Printf("[DIR]  %s (%s)\n", name, sizeFormatted)
		} else {
			fmt.Printf("[FILE] %s (%s)\n", name, sizeFormatted)
		}
	}
	return nil
}

func main() {
	root := flag.String("root", "", "choose a directory")
	sortOrder := flag.String("sort", "asc", "choose sorting of directory (asc/desc)")
	flag.Parse()

	if *root == "" {
		fmt.Println("Please specify a directory using the -root flag.")
		return
	}

	if _, err := os.Stat(*root); os.IsNotExist(err) {
		fmt.Println("Directory does not exist.")
		return
	}

	files, sizes, err := getFilesAndSizes(*root)
	if err != nil {
		fmt.Printf("Error walking the directory: %s\n", err)
		return
	}

	sortFiles(files, sizes, *sortOrder)
	printFiles(files, sizes)
}
