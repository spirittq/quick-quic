package shared

import "encoding/json"


func FromJson[T any](receivedData []byte) (T, error) {
	var decodedData T
	err := json.Unmarshal(receivedData, &decodedData)
	return decodedData, err
}

func ToJson[T any](sendingData T) ([]byte, error) {
	var jsonData []byte
	jsonData, err := json.Marshal(sendingData)
	return jsonData, err
}
