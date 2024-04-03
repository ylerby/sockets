package main

import (
	"bufio"
	"fmt"
	"log"
	"multiprocessing_server/logger"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	InvalidDataMessage = "некорректные данные"
	FullNameSize       = 3
	MaxMessageSize     = 5
	MinMessageSize     = 3
	ServerWriter       = "Сервер написан - Алиякбяров М.А."
	BufferSize         = 256
)

func main() {
	l, err := logger.InitLogger("logs/log_file.txt")
	if err != nil {
		log.Fatalf("failed to initialize application logger: %v", err)
	}

	l.Println("cringe")

	el, err := logger.InitLogger("logs/error_log_file.txt")
	if err != nil {
		log.Fatalf("failed to initialize error logger: %v", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		outputMessage, err := processMessage(input)
		if err != nil {
			el.Printf("error: %v", err)
			continue
		}

		fmt.Println(outputMessage)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "ошибка при чтении: ", err)
	}
}

func processMessage(input string) (string, error) {
	input = strings.TrimSpace(input)
	bufferSlice := strings.Split(input, " ")

	if len(bufferSlice) > MaxMessageSize || len(bufferSlice) < MinMessageSize {
		return "", fmt.Errorf("некорректная длина")
	}

	return InitOutputMessage(input), nil
}

func InitOutputMessage(clientMessage string) string {
	messageSlice := strings.Split(clientMessage, " ")

	result := ""

	for i := FullNameSize - 1; i >= 0; i-- {
		runeWord := []rune(messageSlice[i])
		word := make([]rune, utf8.RuneCountInString(messageSlice[i]))
		for j := utf8.RuneCountInString(messageSlice[i]) - 1; j >= 0; j-- {
			switch j {
			case utf8.RuneCountInString(messageSlice[i]) - 1:
				letter := strings.ToUpper(string(runeWord[j]))
				word[j] = []rune(letter)[0]
			case 0:
				letter := strings.ToLower(string(runeWord[j]))
				word[j] = []rune(letter)[0]
			default:
				word[j] = runeWord[j]
			}
		}
		result += string(word) + " "
	}

	result += ServerWriter + " "

	lastMessagePart := messageSlice[FullNameSize:]
	result += strings.Join(lastMessagePart, " ")

	return result
}
