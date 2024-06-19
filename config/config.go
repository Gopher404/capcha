package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	Bind    *BindConfig    `json:"bind"`
	API     *APIConfig     `json:"api"`
	Service *ServiceConfig `json:"service"`
	Redis   *RedisConfig   `json:"redis"`
	v       string
}

func (c *Config) String() string {
	return c.v
}

type BindConfig struct {
	Ip   string `json:"ip"`
	Port string `json:"port"`
}

type APIConfig struct {
	StrTimeOut string `json:"time_out"`
	TimeOut    time.Duration
}

type ServiceConfig struct {
	Fonts         string `json:"fonts_dir"`
	CapchaWidth   int    `json:"capcha_width"`
	CapchaHeight  int    `json:"capcha_height"`
	FontSize      int    `json:"font_size"`
	LenCapchaText int    `json:"len_capcha_text"`
	CapchaTTL     int    `json:"capcha_ttl"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
}

func GetConfig(path string) (*Config, error) {
	var f *os.File
	if len(path) == 0 {
		// open config in current directory
		var err error
		_, path, _, _ := runtime.Caller(0)
		path = filepath.Dir(path)

		f, err = os.Open(filepath.Join(path, "config.json"))
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		f, err = os.Open(path)
		if err != nil {
			return nil, err
		}
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var cnf Config
	if err := json.Unmarshal(b, &cnf); err != nil {
		return nil, err
	}

	cnf.v = strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(string(b), "\n", ""),
				"\r", ""),
			"\"", "'"),
		" ", "")
	
	to, err := time.ParseDuration(cnf.API.StrTimeOut)
	if err != nil {
		return nil, err
	}
	cnf.API.TimeOut = to
	return &cnf, nil
}
