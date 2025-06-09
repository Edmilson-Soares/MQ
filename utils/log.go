package utils

import (
	"log"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Configura o logger para salvar em arquivo com rotação
func SetupLogger(config LogsConfig) {
	if !config.Enabled {

		return
	}
	log.SetOutput(&lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize, // megabytes
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,   // days
		Compress:   config.Compress, // gzip
	})
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
