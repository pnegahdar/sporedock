package utils

import "encoding/json"

func Marshall(i interface{}) (string, error) {
	resp, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(resp[:]), nil
}

func Unmarshall(data string, i interface{}) error {
	err := json.Unmarshal([]byte(data), i)
	if err != nil {
		return err
	}
	return nil
}
