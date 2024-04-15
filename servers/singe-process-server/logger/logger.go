package logger

import (
	"fmt"
	"log"
	"os"
)

func InitLogger(fileName string) (*log.Logger, error) {
	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии/создании файла - %s\n", err)
	}

	logger := &log.Logger{}
	logger.SetOutput(logFile)

	return logger, nil
}
