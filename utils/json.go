package utils

import (
	"encoding/json"
	"io"
	"strings"
)

func Marshall(i interface{}) (string, error) {
	resp, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func Unmarshall(data string, i interface{}) error {
	err := json.Unmarshal([]byte(data), i)
	if err != nil {
		return err
	}
	return nil
}

func UnmarshalReader(data io.Reader, i interface{}) error {
	err := json.NewDecoder(data).Decode(i)
	if err != nil {
		return err
	}
	return nil
}

func JsonListFromObjects(objects ...string) string {
	return "[" + strings.Join(objects, ", ") + "]"
}
