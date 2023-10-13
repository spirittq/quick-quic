package utils

import "encoding/json"

// Generic function to parse JSON-encoded data into passed parameter
func FromJson[T any](receivedData []byte) (T, error) {
	var decodedData T
	err := json.Unmarshal(receivedData, &decodedData)
	return decodedData, err
}

// Generic function to encode passed parameter into JSON.
func ToJson[T any](sendingData T) ([]byte, error) {
	var jsonData []byte
	jsonData, err := json.Marshal(sendingData)
	return jsonData, err
}
