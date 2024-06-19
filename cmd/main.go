package main

import (
	"capcha/adapters/cache"
	"capcha/adapters/handler"
	"capcha/config"
	"capcha/core/service"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
)

const (
	configPath = "config/config.json"
	logPath    = "logs.json"
)

func main() {
	l, err := InitLogger(logPath)
	if err != nil {
		panic(err)
	}
	l.Info("init logger")

	cnf, err := config.GetConfig(configPath)
	if err != nil {
		l.Error(err.Error())
		return
	}
	l.Info("get config")
	l.Info(cnf.String())

	c, err := cache.NewRedisCache(cnf.Redis.Addr, cnf.Redis.Password)
	if err != nil {
		l.Error(err.Error())
		return
	}
	l.Info("connect to redis")

	s, err := service.New(cnf.Service, c)
	if err != nil {
		l.Error(err.Error())
		return
	}
	l.Info("create service")

	h := handler.New(cnf.API, s, l)
	addr := net.JoinHostPort(cnf.Bind.Ip, cnf.Bind.Port)
	l.Info("listen on " + addr)

	if err := http.ListenAndServe(addr, h); err != nil {
		l.Error(err.Error())
		return
	}

}

func InitLogger(filePath string) (*slog.Logger, error) {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 666)
	if err != nil {
		return nil, err
	}
	w := io.MultiWriter(f, os.Stdout)

	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	l := slog.New(h)

	return l, nil
}
