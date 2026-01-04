package core

import (
	"encoding/json"
	"hexframe/snowfall/schemas"
	"hexframe/snowfall/utils"
)

var logger = utils.RequestLogger()

func ProcessMessage(data []byte) ([]byte, error) {
	logger.Debug("Processing one...")
	var decodedMessage schemas.Message
	err := json.Unmarshal(data, &decodedMessage)
	if err != nil {
		return data, err
	}
	logger.Debug("Message:", "method", decodedMessage.Method)
	return data, nil
}
