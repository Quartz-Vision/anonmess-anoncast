package settings

import (
	"strconv"

	"github.com/Quartz-Vision/golog"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

var Config = struct {
	ServerPort int    `env:"SERVER_PORT" envDefault:"8000"`
	ServerHost string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	ServerAddr string
}{}

func init() {
	godotenv.Load()

	if err := env.Parse(&Config); err != nil {
		golog.Error.Fatalln(err.Error())
	}

	Config.ServerAddr = Config.ServerHost + ":" + strconv.FormatInt(int64(Config.ServerPort), 10)
}
