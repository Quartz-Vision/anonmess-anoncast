package settings

import (
	"errors"
	"strconv"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

var ErrEnvLoading = errors.New("error while loading .env file")
var ErrEnvParsing = errors.New("error while parsing env")

var Config = struct {
	ServerPort int `env:"SERVER_PORT" envDefault:"8000"`
	ServerHost string `env:"SERVER_HOST" envDefault:"127.0.0.1"`
	ServerAddr string
} {}

func Init() error {
	if err := godotenv.Load(); err != nil {
		return ErrEnvLoading
	}

	if err := env.Parse(&Config); err != nil {
		return ErrEnvParsing
	}
	
	Config.ServerAddr = Config.ServerHost + ":" + strconv.FormatInt(int64(Config.ServerPort), 10)

	return nil
}
