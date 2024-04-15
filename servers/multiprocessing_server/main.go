package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"multiprocessing_server/logger"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	protocol    = "tcp"
	connhandler = "./connection-handler"
)

type Service struct {
	l   *log.Logger
	el  *log.Logger
	cfg *Config
}

func New() *Service {
	return new(Service)
}

type Config struct {
	Address          string `json:"address"`
	Port             string `json:"port"`
	LogFileName      string `json:"log_file_name"`
	ErrorLogFileName string `json:"error_log_file_name"`
	Timeout          int    `json:"timeout"`
}

func (s *Service) handleConnection(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}(conn)

	command := exec.Command(connhandler)
	command.Stdin = conn
	command.Stdout = conn
	command.Stderr = os.Stderr

	errCh := make(chan error, 1)
	disconnectTicker := time.NewTicker(time.Second * 20 * time.Duration(s.cfg.Timeout))
	defer disconnectTicker.Stop()

	go func() {
		err := command.Start()
		if err != nil {
			s.el.Printf("failed to start command: %v", err)
			errCh <- err
			return
		}

		err = command.Wait()
		errCh <- err
	}()

	select {
	case <-disconnectTicker.C:
		if command.Process != nil {
			err := command.Process.Kill()
			if err != nil {
				s.el.Printf("failed to kill process: %v", err)
				return
			}
		}
	case err := <-errCh:
		if err != nil {
			s.el.Printf("command finished with error: %v", err)
		}
	}
}

func main() {
	cfgfilename := flag.String("cfgfilepath", "config/config.json", "")
	logfilepath := flag.String("logfilepath", "logs/log_file.txt", "output log file path")
	errlogfilepath := flag.String("errlogfilepath", "logs/error_log_file.txt", "output error log file path")
	flag.Parse()

	service := New()

	var err error

	service.cfg, err = service.InitConfig(*cfgfilename)
	if err != nil {
		service.el.Fatal(err)
	}

	service.l, err = logger.Init(*logfilepath)
	if err != nil {
		log.Fatalf("failed to init service l: %v", err)
	}

	service.el, err = logger.Init(*errlogfilepath)
	if err != nil {
		log.Fatalf("failed to initialize error l: %v", err)
	}

	server, err := net.Listen(protocol, service.cfg.Address+":"+service.cfg.Port)
	if err != nil {
		service.el.Fatalf("failed to start server: %v", err)
	}

	defer func(server net.Listener) {
		err := server.Close()
		if err != nil {
			return
		}
	}(server)

	service.l.Printf("%v: Сервер запущен: %s", time.Now(), service.cfg.Address+":"+service.cfg.Port)

	wg := new(sync.WaitGroup)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-ctx.Done()
		err := server.Close()
		if err != nil {
			return
		}

		os.Exit(0)
	}()

	for {
		conn, err := server.Accept()
		if err != nil {
			service.el.Printf("connection error: %v\n", err)
			continue
		}

		service.l.Printf("%v: Клиент подключен", time.Now())

		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			service.handleConnection(conn, wg)
		}(wg)

	}

	wg.Wait()
}

func (s *Service) InitConfig(cfgfilename string) (*Config, error) {
	file, err := os.Open(cfgfilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open cfg file: %w\n", err)
	}

	cfg := new(Config)

	decoder := json.NewDecoder(file)

	err = decoder.Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("error occurred while data load: %w\n", err)
	}

	return cfg, nil
}
