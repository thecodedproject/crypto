package io

import (
	"encoding/json"
	"io/ioutil"
)

func UnmarshalJsonFile(path string, i interface{}) error {

	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonFile, &i)
	if err != nil {
		return err
	}

	return nil
}
