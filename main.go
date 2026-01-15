package main

import (
	"bytes"
	"hexframe/snowfall/core"
	"hexframe/snowfall/schemas"
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

type ConfigLog struct {
	Mode     string `validate:"oneof=both stderr file"`
	Filename string `validate:"required_unless=Mode stderr,filepath"`
	Level    string `validate:"oneof=fatal error warn info debug"`
}
type Config struct {
	Host          string   `validate:"hostname_port"`
	InCheck       string   `validate:"oneof=on off"`
	InTransport   string   `validate:"oneof=http stdio"`
	OutCheck      string   `validate:"oneof=on off"`
	OutCommand    []string `validate:"required_if=OutCheck on OutTransport stdio,dive"`
	OutEndpoint   string   `validate:"required_if=OutCheck on OutTransport http,url"`
	OutTransport  string   `validate:"oneof=http stdio"`
	Audit         string   `validate:"oneof=on off"`
	AuditEndpoint string   `validate:"required_if=Audit on,url"`
	AuditTreshold int      `validate:"required_if=Audit on,number,min=0,max=100"`
}

var config = Config{}
var logger = utils.Logger
var loggerConfig = ConfigLog{}

func main() {
	// Инициализируем валидатор
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Загружаем .env файл
	err := godotenv.Load()

	// Читаем настройки для логгера
	loggerConfig.Level = utils.Getenv("SF_LOG_LEVEL", "info")
	loggerConfig.Mode = utils.Getenv("SF_LOG_MODE", "stderr")
	loggerConfig.Filename = os.Getenv("SF_LOG_FILENAME")
	err = validate.Struct(loggerConfig)
	if err != nil {
		validationErrors, exists := err.(validator.ValidationErrors)
		if exists {
			logger.Fatal(
				"Validation failed:",
				"field", validationErrors[0].Field(),
				"value", validationErrors[0].Value(),
				"tag", validationErrors[0].Tag(),
			)
		}
	}

	// Подготавливаем логгер к работе
	utils.PrepareLogger()

	// Настраиваем уровень логгирования
	if loggerConfig.Level != "info" {
		utils.Logger.SetLevel(schemas.LogLevels[loggerConfig.Level])
	}
	// Настраиваем вывод логгера
	var logWriters []io.Writer
	if loggerConfig.Mode == "stderr" || loggerConfig.Mode == "both" {
		logWriters = append(logWriters, os.Stderr)
	}
	if loggerConfig.Mode == "file" || loggerConfig.Mode == "both" {
		f, err := os.OpenFile(
			loggerConfig.Filename,
			os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600,
		)
		if err != nil {
			logger.Fatal("Could not open file for logging")
		}
		logWriters = append(logWriters, f)
	}
	// Проверяем количество установленных выводов.
	// Необходимо для сохранения отображения цветов в консоли
	if len(logWriters) == 1 {
		utils.Logger.SetOutput(logWriters[0])
	} else {
		utils.Logger.SetOutput(io.MultiWriter(logWriters...))
	}

	// Записываем значения из переменных окружения в структуру с настройками
	logger.Info("Configuration: Processing environment variables...")
	config.Host = utils.Getenv("SF_HOST", "127.0.0.1:8080")
	config.InCheck = utils.Getenv("SF_IN_CHECK", "on")
	config.InTransport = os.Getenv("SF_IN_TRANSPORT")
	config.OutCheck = utils.Getenv("SF_OUT_CHECK", "on")
	config.OutCommand = strings.Split(os.Getenv("SF_OUT_CMD"), " ")
	config.OutEndpoint = os.Getenv("SF_OUT_URL")
	config.OutTransport = os.Getenv("SF_OUT_TRANSPORT")
	config.Audit = utils.Getenv("SF_AUDIT", "off")
	config.AuditEndpoint = os.Getenv("SF_AUDIT_URL")
	config.AuditTreshold, err = strconv.Atoi(utils.Getenv("SF_AUDIT_TRESHOLD", "80"))
	if err != nil {
		logger.Fatal(
			"Read failed:",
			"field", "SF_AUDIT_TRESHOLD",
			"value", os.Getenv("SF_AUDIT_TRESHOLD"),
			"reason", "INCORRECT_NUMBER",
		)
	}

	// Проверяем переданные параметры на корректность
	logger.Info("Configuration: Validating the provided parameters...")
	err = validate.Struct(config)
	if err != nil {
		validationErrors, exists := err.(validator.ValidationErrors)
		if exists {
			logger.Fatal(
				"Validation failed:",
				"field", validationErrors[0].Field(),
				"value", validationErrors[0].Value(),
				"tag", validationErrors[0].Tag(),
			)
		}
	}

	// В зависимости от транспортов могут быть использованы:
	// - Прокси-сервер (HTTP -> HTTP)
	// - HTTP-клиент (STDIO -> HTTP)
	// - HTTP-сервер (HTTP -> STDIO)
	// - Только Stdin и Stdout (STDIO -> STDIO)
	if config.InTransport == "stdio" && config.OutTransport == "stdio" {
		// TODO: Implement
	} else {
		serverURL, _ := url.Parse(config.OutEndpoint)
		proxy := &httputil.ReverseProxy{}
		proxy.Director = func(req *http.Request) {
			if config.InCheck == "on" && req.Method == http.MethodPost && req.Body != nil {
				bodyBytes, _ := io.ReadAll(req.Body)
				if len(bodyBytes) > 0 {
					bodyBytes, err = core.ProcessRequest(bodyBytes)
					if err != nil {
						logger.Error(err)
					}
				}
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
			req.URL = serverURL
		}
		proxy.ModifyResponse = func(resp *http.Response) error {
			if config.OutCheck == "on" && resp.Request.Method == http.MethodPost && resp.Body != nil {
				bodyBytes, _ := io.ReadAll(resp.Body)
				if len(bodyBytes) > 0 {
					bodyBytes, err = core.ProcessResponse(bodyBytes)
					if err != nil {
						logger.Error(err)
					}
				}
				resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
			return nil
		}
		logger.Fatal(
			http.ListenAndServe(config.Host, proxy),
		)
	}
}
