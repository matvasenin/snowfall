package core

import (
	"bufio"
	"bytes"
	"encoding/json"
	"hexframe/snowfall/schemas"
	"hexframe/snowfall/utils"
	"strings"

	"github.com/go-playground/validator/v10"
)

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
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(messageData)
	if err != nil {
		return data, err
	}
	if messageData.Error.Message != "" {
		logger.Warn("MCP Error:", "message", messageData.Error.Message)
	}
	if messageData.Result.Tools != nil {
		for index, tool := range messageData.Result.Tools {
			logger.Debug(
				"Tool:",
				"index", index,
				"name", tool.Name,
			)
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
	return data, nil
}
