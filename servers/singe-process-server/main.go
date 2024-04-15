package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"server/logger"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	LogFileName        = "logs/log_file.txt"
	ErrorLogFileName   = "logs/error_log_file.txt"
	ConfigFileName     = "config/config.json"
	ServerProtocol     = "tcp"
	ServerWriter       = "Сервер написан - Алиякбяров М.А."
	InvalidDataMessage = "некорректные данные"
	BufferSize         = 256
	FullNameSize       = 3
	MaxMessageSize     = 5
	MinMessageSize     = 3
)

type Application struct {
	logger      *log.Logger
	errorLogger *log.Logger
}

func NewApplication() *Application {
	return &Application{}
}

type Config struct {
	Address          string `json:"address"`
	Port             string `json:"port"`
	LogFileName      string `json:"log_file_name"`
	ErrorLogFileName string `json:"error_log_file_name"`
	Timeout          int    `json:"timeout"`
}

func (a *Application) handleConnection(conn net.Conn, timeout int) {
	defer func(conn net.Conn) {
		err := conn.Close()
		a.logger.Printf("%s - клиент отключен\n", time.Now())
		if err != nil {
			a.errorLogger.Fatalf("ошибка при закрытии соединения с клиентом - %s\n", err)
		}
	}(conn)

	time.Sleep(time.Second * time.Duration(timeout))

	buffer := make([]byte, BufferSize)
	n, err := conn.Read(buffer)
	if err != nil {
		a.errorLogger.Printf("ошибка при получении сообщения - %s\n", err)
		return
	}

	a.logger.Printf("%s - получено сообщение от клиента - %s", time.Now(), string(buffer[:n]))

	bufferSlice := strings.Split(string(buffer), " ")
	if len(bufferSlice) > MaxMessageSize || len(bufferSlice) < MinMessageSize {
		_, err = conn.Write([]byte(InvalidDataMessage))
		a.logger.Printf("%s - отправлено сообщение клиенту - %s\n", time.Now(), InvalidDataMessage)
		if err != nil {
			a.errorLogger.Fatalf("ошибка при ответе - %s\n", err)
		}
		return
	}

	time.Sleep(time.Second * time.Duration(timeout))

	outputMessage := InitOutputMessage(string(buffer), n)

	_, err = conn.Write([]byte(outputMessage))
	if err != nil {
		a.errorLogger.Fatalf("ошибка при ответе - %s", err)
	}

	a.logger.Printf("%s - отправлено сообщение клиенту - %s\n", time.Now(), outputMessage)
}

func main() {
	var err error

	app := NewApplication()
	app.logger, err = logger.InitLogger(LogFileName)
	if err != nil {
		log.Fatalf("ошибка при инициализации логгера - %s\n", err)
	}

	app.errorLogger, err = logger.InitLogger(ErrorLogFileName)
	if err != nil {
		log.Fatalf("ошибка при инициализации логгера - %s\n", err)
	}

	config, err := app.InitConfig()
	if err != nil {
		app.errorLogger.Fatalf("ошибка при получении конфига сервера - %s\n", err)
	}

	server, err := net.Listen(ServerProtocol, fmt.Sprintf("%s:%s",
		config.Address, config.Port))
	if err != nil {
		app.errorLogger.Fatalf("ошибка при запуске сервера - %s\n", err)
	}

	app.logger.Printf("%s - сервер запущен", time.Now())

	defer func(server net.Listener) {
		err := server.Close()
		if err != nil {
			app.errorLogger.Printf("ошибка при завершении работы сервера - %s\n", err)
		}
	}(server)

	for {
		connection, err := server.Accept()
		if err != nil {
			app.errorLogger.Printf("ошибка при соединении - %s\n", err)
			continue
		}

		app.logger.Printf("%s - клиент подключен", time.Now())

		app.handleConnection(connection, config.Timeout)
	}
}

func (a *Application) InitConfig() (*Config, error) {
	file, err := os.Open(ConfigFileName)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии файла - %s\n", err)
	}

	currentServerConfig := &Config{}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(currentServerConfig)
	if err != nil {
		return nil, fmt.Errorf("ошибка при десериализации - %s\n", err)
	}

	return currentServerConfig, nil
}

func InitOutputMessage(clientMessage string, byteNums int) string {
	clientMessage = clientMessage[:byteNums]
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
