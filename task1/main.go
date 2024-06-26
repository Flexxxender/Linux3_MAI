package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func main() {
	// Проверяем, был ли предоставлен путь к файлу в качестве аргумента командной строки
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <filename>")
		os.Exit(1)
	}

	// Получаем путь к файлу из аргументов командной строки
	filePath := os.Args[1]

	// Открываем файл для чтения
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Создаем новый процесс для вывода данных
	cmd := exec.Command("cat") // Пример использования команды cat для вывода содержимого файла
	cmd.Stdout = os.Stdout     // Направляем стандартный вывод процесса в стандартный вывод текущей программы

	// Запускаем процесс
	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting command:", err)
		os.Exit(1)
	}

	// Копируем содержимое файла в канал Pipe (в стандартный ввод процесса)
	_, err = io.Copy(cmd.Stdout, file)
	if err != nil {
		fmt.Println("Error writing to pipe:", err)
		os.Exit(1)
	}

	// Ждем завершения процесса
	err = cmd.Wait()
	if err != nil {
		fmt.Println("Error waiting for command:", err)
		os.Exit(1)
	}
}
