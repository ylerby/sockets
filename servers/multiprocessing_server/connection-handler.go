package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"multiprocessing_server/logger"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	InvalidDataMessage = "некорректные данные"
	FullNameSize       = 3
	MaxMessageSize     = 5
	MinMessageSize     = 3
	ServerWriter       = "Сервер написан - Алиякбяров М.А."
)

type Subservice struct {
	l   *log.Logger
	el  *log.Logger
	cfg *Subconfig
}

type Subconfig struct {
	Address          string `json:"address"`
	Port             string `json:"port"`
	LogFileName      string `json:"log_file_name"`
	ErrorLogFileName string `json:"error_log_file_name"`
	Timeout          int    `json:"timeout"`
}

func main() {
	cfgfilename := "config/config.json"
	logfilepath := "logs/log_file.txt"
	flag.Parse()

	subservice := new(Subservice)

	cfg, err := subservice.InitConfig(cfgfilename)
	if err != nil {
		subservice.el.Fatal(err)
	}

	l, err := logger.Init(logfilepath)
	if err != nil {
		log.Fatalf("failed to initialize application l: %v", err)
	}

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		input := sc.Text()

		l.Printf("%v: Получено сообщение от клиента: %s", time.Now(), input)

		outputmsg, err := InitOutputMessage(input)
		if err != nil {
			time.Sleep(time.Second * time.Duration(cfg.Timeout))
			fmt.Println(InvalidDataMessage)
			break
		}

		time.Sleep(time.Second * time.Duration(cfg.Timeout))

		fmt.Println(outputmsg)
		l.Printf("%v: Отправлено сообщение клиенту: %s", time.Now(), outputmsg)

		time.Sleep(time.Second * time.Duration(cfg.Timeout))
		l.Printf("%s - Клиент отключен", time.Now())
	}

	if err := sc.Err(); err != nil {
		_, err := fmt.Fprintln(os.Stdout, "failed to read: ", err)
		if err != nil {
			log.Fatalf("failed to reading")
		}
	}
}

func InitOutputMessage(clientMessage string) (string, error) {
	clientMessage = strings.TrimSpace(clientMessage)
	bufferSlice := strings.Split(clientMessage, " ")

	if len(bufferSlice) > MaxMessageSize || len(bufferSlice) < MinMessageSize {
		return "", fmt.Errorf("некорректная длина")
	}

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

	return result, nil
}

func (s *Subservice) InitConfig(cfgfilename string) (*Subconfig, error) {
	file, err := os.Open(cfgfilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open cfg file: %w\n", err)
	}

	cfg := new(Subconfig)

	decoder := json.NewDecoder(file)

	err = decoder.Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("error occurred while data load: %w\n", err)
	}

	return cfg, nil
}
