package main

import (
	"encoding/json"
	"fmt"
	"log"
	"multiprocessing_server/logger"
	"net"
	"os"
	"os/exec"
	"sync"
)

const (
	ConfigFileName = "config/config.json"
	ServerProtocol = "tcp"
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

func (a *Application) handleConnection(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(conn)

	cmd := exec.Command("./connection-handler")
	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		a.errorLogger.Printf("ошибка при запуске процесса: %s", err)
		return
	}

	if err := cmd.Wait(); err != nil {
		a.errorLogger.Printf("ошибка в процессе обработки: %s", err)
	}
}

func main() {
	app := NewApplication()

	config, err := app.InitConfig()
	if err != nil {
		app.errorLogger.Fatal(err)
	}

	app.logger, err = logger.InitLogger("logs/log_file.txt")
	if err != nil {
		log.Fatalf("failed to initialize application logger: %v", err)
	}

	app.errorLogger, err = logger.InitLogger("logs/error_log_file.txt")
	if err != nil {
		log.Fatalf("failed to initialize error logger: %v", err)
	}

	server, err := net.Listen(ServerProtocol, config.Address+":"+config.Port)
	if err != nil {
		app.errorLogger.Fatalf("ошибка при запуске сервера: %s", err)
	}

	defer func(server net.Listener) {
		err := server.Close()
		if err != nil {
			return
		}
	}(server)

	app.logger.Printf("сервер запущен на %s", config.Address+":"+config.Port)

	var wg sync.WaitGroup

	for {
		conn, err := server.Accept()
		if err != nil {
			app.errorLogger.Printf("ошибка при соединении: %s", err)
			continue
		}
		wg.Add(1)
		go app.handleConnection(conn, &wg)
	}

	wg.Wait()
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
