package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Функция для рекурсивного обхода директории и сбора информации о файлах и папках
func getFilesAndSizes(root string) ([]string, []int64, error) {
	var files []string
	var sizes []int64 

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Игнорируем корневую директорию
		if path == root {
			return nil
		}

		// Добавляем информацию о файле/папке
		files = append(files, path)
		if info.IsDir() {
			size, err := getDirSize(path)
			if err != nil {
				return err
			}
			sizes = append(sizes, size)
		} else {
			sizes = append(sizes, info.Size())
		}

		return nil
	})

	return files, sizes, err
}
	
// Функция для вычисления размера директории
func getDirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		} 
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// Функция для сортировки файлов и папок по размеру
func sortFiles(files []string, sizes []int64, order string) {
	// Создаем срез пар (путь, размер) для сортировки
	type fileSize struct {
		path string
		size int64
	}
	var fileSizes []fileSize
	for i := range files {
		fileSizes = append(fileSizes, fileSize{files[i], sizes[i]})
	}

	// Сортируем
	sort.Slice(fileSizes, func(i, j int) bool {
		if order == "asc" {
			return fileSizes[i].size < fileSizes[j].size
		} else {
			return fileSizes[i].size > fileSizes[j].size
		}
	})

	// Обновляем исходные срезы
	for i := range fileSizes {
		files[i] = fileSizes[i].path
		sizes[i] = fileSizes[i].size
	}
}

// Функция для конвертации размера в удобочитаемый формат
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

// Функция для вывода информации о файлах и папках
func printFiles(files []string, sizes []int64) {
	for i, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			fmt.Printf("Error getting info for %s: %s\n", file, err)
			continue
		}

		// Извлекаем только имя файла/папки
		name := filepath.Base(file)

		// Форматируем размер
		sizeFormatted := formatSize(sizes[i])

		// Выводим информацию
		if info.IsDir() {
			fmt.Printf("[DIR]  %s (%s)\n", name, sizeFormatted)
		} else {
			fmt.Printf("[FILE] %s (%s)\n", name, sizeFormatted)
		}
	}
}

func main() {
	
	root := flag.String("root", "", "choose a directory")
	sortOrder := flag.String("sort", "asc", "choose sorting of directory (asc/desc)")
	flag.Parse()

	// Проверка, что директория указана
	if *root == "" {
		fmt.Println("Please specify a directory using the -root flag.")
		return
	}

	// Проверка, что директория существует
	if _, err := os.Stat(*root); os.IsNotExist(err) {
		fmt.Println("Directory does not exist.")
		return
	}

	// Получаем список файлов и папок
	files, sizes, err := getFilesAndSizes(*root)
	if err != nil {
		fmt.Printf("Error walking the directory: %s\n", err)
		return
	}

	// Сортируем файлы и папки по размеру
	sortFiles(files, sizes, *sortOrder)

	// Выводим результат
	printFiles(files, sizes)
}