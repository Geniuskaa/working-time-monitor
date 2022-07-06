package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func main() {
}

func logger() *zap.SugaredLogger {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	encore := zapcore.NewJSONEncoder(encoderConfig)
	file, err := os.Create("./internal/log/logs.txt")
	if err != nil {
		panic("Error with creating file")
	}
	writeSyncer := zapcore.AddSync(file)
	core := zapcore.NewCore(encore, writeSyncer, zapcore.ErrorLevel)

	sugarLogger := zap.New(core).Sugar()

	return sugarLogger
}
