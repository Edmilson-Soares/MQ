package utils

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type ServerConfig struct {
	//HTTP HTTPConfig `toml:"http"`
	MQ MQConfig `toml:"mq"`
	//MQTT MQTTConfig `toml:"mqtt"`
	Logs LogsConfig `toml:"logs"`
	//Proc ProcConfig `toml:"proc"`
}

type HTTPConfig struct {
	Enabled  bool   `toml:"enabled"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	UseHTTPS bool   `toml:"use_https"`
	CertFile string `toml:"cert_file,omitempty"`
	KeyFile  string `toml:"key_file,omitempty"`
}
type MQTTConfig struct {
	Enabled bool `toml:"enabled"`

	Broker       string        `toml:"broker"`
	Port         int           `toml:"port"`
	KeepAlive    time.Duration `toml:"keep_alive"`
	CleanSession bool          `toml:"clean_session"`
	QoS          int           `toml:"qos"`
}
type MQConfig struct {
	Enabled  bool   `toml:"enabled"`
	Broker   string `toml:"broker"`
	FileKV   string `toml:"kvfile"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type User struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	IsAdmin  bool   `toml:"is_admin"`
}

type LogsConfig struct {
	Enabled    bool   `toml:"enabled"`
	Filename   string `toml:"filename"`
	MaxSize    int    `toml:"maxSize"`    // em megabytes
	MaxBackups int    `toml:"maxBackups"` // quantidade de backups
	MaxAge     int    `toml:"maxAge"`     // em dias
	Compress   bool   `toml:"compress"`   // compactar backups
}

type ProcConfig struct {
	MaxProc int `toml:"maxProc"` // número máximo de processos
}

func GetConfig() (*ServerConfig, error) {
	configFile := "config.toml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	// Lê o arquivo de configuração
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// Decodifica o TOML
	var config ServerConfig
	err = toml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	SetupLogger(config.Logs)
	return &config, nil
}
