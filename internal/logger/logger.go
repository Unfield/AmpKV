package logger

import (
	"log"

	"go.uber.org/zap"
)

var baseLogger *zap.Logger

func InitLogger() {
	var err error
	baseLogger, err = zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize zap: %v", err)
	}
	zap.ReplaceGlobals(baseLogger)
}

func GetLogger() *zap.Logger {
	if baseLogger == nil {
		InitLogger()
		if baseLogger == nil {
			log.Fatalf("failed to initialize zap")
		}
	}
	return baseLogger
}
