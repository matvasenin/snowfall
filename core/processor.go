package core

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hexframe/snowfall/schemas"
	"hexframe/snowfall/utils"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ErrorHelper struct {
	Codename string
	Reason   string
}

var errorHelpers = map[int]ErrorHelper{
	-32000: {
		Codename: "SERVER_ERROR",
		Reason:   "Server-side tool execution failed",
	},
	-32001: {
		Codename: "RESOURCE_NOT_FOUND",
		Reason:   "The requested resource could not be found",
	},
	-32002: {
		Codename: "PERMISSION_DENIED",
		Reason:   "The client lacks the necessary permissions",
	},
	-32003: {
		Codename: "RATE_LIMIT",
		Reason:   "Rate limit of server was exceeded",
	},
	-32004: {
		Codename: "TIMEOUT",
		Reason:   "The operation timed out",
	},
	-32600: {
		Codename: "METHOD_NOT_FOUND",
		Reason:   "The requested operation is not available",
	},
	-32602: {
		Codename: "INVALID_PARAM",
		Reason:   "Invalid parameter was passed by client",
	},
	-32603: {
		Codename: "INTERNAL_ERROR",
		Reason:   "Server encountered on implementation issue",
	},
	-32700: {
		Codename: "PARSE_ERROR",
		Reason:   "MCP Message is malformed",
	},
	-32800: {
		Codename: "REQUEST_CANCELLED",
		Reason:   "Client have cancelled the request",
	},
	-32801: {
		Codename: "CONTENT_TOO_LARGE",
		Reason:   "The content of the request is too large",
	},
}

var logger = utils.Logger

func ProcessResponse(data []byte) ([]byte, error) {
	var messageData schemas.MCPResponse
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		key, value, exists := strings.Cut(scanner.Text(), ": ")
		if exists && key == "data" {
			err := json.Unmarshal([]byte(value), &messageData)
			if err != nil {
				return data, err
			}
			break
		}
	}
	logger.Debug("Processing response:", "id", messageData.ID)
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(messageData)
	if err != nil {
		return data, err
	}
	if messageData.Error.Message != "" {
		helper, exists := errorHelpers[messageData.Error.Code]
		if exists {
			logger.Warn(
				"Known MCP Error:",
				"code", messageData.Error.Code,
				"codename", helper.Codename,
				"reason", helper.Reason,
			)
		} else {
			logger.Warn(
				"Unknown MCP Error:",
				"code", messageData.Error.Code,
				"message", messageData.Error.Message,
			)
		}
	}
	if messageData.Result.Tools != nil {
		for index, tool := range messageData.Result.Tools {
			logger.Debug(
				"Tool was discovered:",
				"index", index,
				"name", tool.Name,
			)
			if tool.Description != "" {
				logger.Debug("Checking description of the tool...")
				if utils.Config.Audit == "on" {
					logger.Debug("Audit Service is present. Sending request...")
					jsonPayload, _ := json.Marshal(
						map[string]string{
							"token": utils.Config.AuditToken,
							"text":  tool.Description,
						},
					)
					resp, err := http.Post(
						fmt.Sprintf("%s/check", utils.Config.AuditEndpoint),
						"application/json", bytes.NewBuffer(jsonPayload),
					)
					if err != nil {
						logger.Fatal(
							"Connection error:",
							"message", err,
						)
					}
					defer resp.Body.Close()
					scoreRaw, _ := io.ReadAll(resp.Body)
					score, err := strconv.ParseInt(string(scoreRaw), 10, 8)
					if err != nil {
						logger.Fatal("Convertion failed:", "message", err)
					}
					if int(score) > utils.Config.AuditThreshold {
						logger.Warn(
							fmt.Sprintf("Description score is %d: Injection detected", score),
						)
						return data, errors.New("Injection Detected")
					}
					logger.Debug(
						fmt.Sprintf("Description score is %d: Looks good", score),
					)
				}
				// TODO: Фильтры
			}
		}
	}
	return data, nil
}

func ProcessRequest(data []byte) ([]byte, error) {
	var messageData schemas.MCPRequest
	err := json.Unmarshal(data, &messageData)
	if err != nil {
		return data, err
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(messageData)
	if err != nil {
		return data, err
	}
	logger.Debug(
		"Processing request:",
		"id", messageData.ID,
		"method", messageData.Method,
	)
	return data, nil
}
