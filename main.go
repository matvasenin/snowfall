package main

import (
	"bytes"
	"fmt"
	"hexframe/snowfall/core"
	"hexframe/snowfall/schemas"
	"hexframe/snowfall/utils"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var config = utils.Config
var logger = utils.Logger
var loggerConfig = schemas.ConfigLog{}

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
			logger.Fatal("Could not create file for logging")
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
	logger.Info("Config: Reading the environment variables...")
	config.Host = utils.Getenv("SF_HOST", "127.0.0.1:8080")
	config.InCheck = utils.Getenv("SF_IN_CHECK", "on")
	config.InTransport = os.Getenv("SF_IN_TRANSPORT")
	config.OutCheck = utils.Getenv("SF_OUT_CHECK", "on")
	config.OutTransport = os.Getenv("SF_OUT_TRANSPORT")
	config.OutCommand = strings.Split(os.Getenv("SF_OUT_CMD"), " ")
	config.OutEndpoint = os.Getenv("SF_OUT_URL")
	config.Audit = utils.Getenv("SF_AUDIT", "off")
	config.AuditEndpoint = os.Getenv("SF_AUDIT_URL")
	config.AuditThreshold, err = strconv.Atoi(utils.Getenv("SF_AUDIT_THRESHOLD", "80"))
	if err != nil {
		logger.Fatal(
			"Parsing failed:",
			"field", "SF_AUDIT_THRESHOLD",
			"value", os.Getenv("SF_AUDIT_THRESHOLD"),
			"reason", "NOT_A_NUMBER",
		)
	}
	logger.Info("Config: Environment variables were loaded.")

	// Проверяем переданные параметры на корректность
	logger.Info("Config: Validating the provided parameters...")
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
	if config.InTransport == "stdio" && config.OutTransport == "stdio" {
		logger.Fatal(
			"Checkup failed:",
			"reason", "MULTIPLE_STDIO",
			"message", "Standard I/O can be used only on one side.",
		)
	}
	logger.Info(
		fmt.Sprintf("Config: All %d parameters are valid.", reflect.TypeFor[schemas.Config]().NumField()),
	)

	// В зависимости от транспортов могут быть использованы:
	// - Прокси-сервер (HTTP -> HTTP);
	// - HTTP-клиент (STDIO -> HTTP);
	// - HTTP-сервер (HTTP -> STDIO).
	// Заметьте, STDIO -> STDIO быть не должно.
	if config.InTransport == "http" && config.OutTransport == "http" {
		logger.Info("Startup: Initilizing proxy server...")
		serverURL, _ := url.Parse(config.OutEndpoint)
		proxy := &httputil.ReverseProxy{}
		// Настраиваем перехватчик сообщений вида "Запрос" (Request)
		proxy.Director = func(req *http.Request) {
			// Инициализруем проверку при определённых условиях:
			// - Она включена для входа;
			// - Тип HTTP-запроса - POST;
			// - Запрос несёт полезную нагрузку - "Тело" (Body).
			if config.InCheck == "on" && req.Method == http.MethodPost && req.Body != nil {
				// Читаем тело в виде байтов
				bodyBytes, _ := io.ReadAll(req.Body)
				// Проверяем на сущестование извлечённых байтов
				if len(bodyBytes) > 0 {
					// Обрабатываем сообщение
					bodyBytes, err = core.ProcessRequest(bodyBytes)
					if err != nil {
						// TODO: Внедрить обработку запроса при отрицательном результате
						logger.Error(err)
					}
				}
				// Возвращаем данные на место
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
			req.URL = serverURL
		}
		// Настраиваем перехватчик сообщений вида "Ответ" (Response)
		proxy.ModifyResponse = func(resp *http.Response) error {
			// Инициализруем проверку при определённых условиях:
			// - Она включена для выхода;
			// - Тип HTTP-запроса - POST;
			// - Ответ несёт полезную нагрузку - "Тело" (Body).
			if config.OutCheck == "on" && resp.Request.Method == http.MethodPost && resp.Body != nil {
				// Читаем тело в виде байтов
				bodyBytes, _ := io.ReadAll(resp.Body)
				// Проверяем на сущестование извлечённых байтов
				if len(bodyBytes) > 0 {
					// Обрабатываем сообщение
					bodyBytes, err = core.ProcessResponse(bodyBytes)
					if err != nil {
						// TODO: Внедрить обработку ответа при отрицательном результате
						logger.Error(err)
					}
				}
				// Возвращаем данные на место
				resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
			return nil
		}
		logger.Info("Startup: Server was initialized successfully.")
		logger.Info("Startup: Launching the proxy server...")
		err = http.ListenAndServe(config.Host, proxy)
		if err != nil {
			logger.Fatal("Proxy crashed:", "error", err)
		}
	} else if config.InTransport == "stdio" {
		// TODO: Implement plain HTTP Server
	} else if config.OutTransport == "stdio" {
		// TODO: Implement plain HTTP Client
	}
}
