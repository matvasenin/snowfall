package main

import (
	"bytes"
	"hexframe/snowfall/core"
	"hexframe/snowfall/utils"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerTransport string   `validate:"oneof=http stdio"`
	ServerEndpoint  string   `validate:"required_without=ServerCommand|url"`
	ServerCommand   []string `validate:"required_without=ServerEndpoint,dive"`
	ClientTransport string   `validate:"oneof=http stdio"`
}

var config = Config{}
var logger = utils.RequestLogger()

func main() {
	// Загружаем .env файл
	logger.Info("Configuration (1/3): Loading .env file...")
	err := godotenv.Load()
	if err != nil {
		logger.Error(err)
		logger.Warn("Configuration (1/3): Could not load .env file")
	}

	// Записываем значения из переменных окружения в структуру с настройками
	logger.Info("Configuration (2/3): Reading the environment variables...")
	config.ServerTransport = os.Getenv("SF_SERVER_TRANSPORT")
	config.ServerEndpoint = os.Getenv("SF_SERVER_URL")
	config.ServerCommand = strings.Split(os.Getenv("SF_SERVER_CMD"), " ")
	config.ClientTransport = os.Getenv("SF_CLIENT_TRANSPORT")

	// Проверяем переданные параметры на корректность
	logger.Info("Configuration (3/3): Validating configuration parameters...")
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(config)
	if err != nil {
		logger.Fatal(err)
	}

	serverURL, _ := url.Parse(config.ServerEndpoint)
	proxy := httputil.NewSingleHostReverseProxy(serverURL)
	proxy.ModifyResponse = func(resp *http.Response) error {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		body, err = core.ProcessMessage(body)
		if err != nil {
			return err
		}

		resp.Body = io.NopCloser(bytes.NewReader(body))
		resp.ContentLength = int64(len(body))
		resp.Header.Set("Content-Length", strconv.Itoa(len(body)))
		return nil
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Body != nil && req.Method == http.MethodPost {
			bodyBytes, _ := io.ReadAll(req.Body)
			bodyBytes, err = core.ProcessMessage(bodyBytes)
			if err != nil {
				logger.Error(err)
				return
			}
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
		proxy.ServeHTTP(w, req)
	})
	logger.Fatal(http.ListenAndServe(":8080", handler))
}
