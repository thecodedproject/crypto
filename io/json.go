package io

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type AuthConfig struct {
	ApiKey string `json:"api_key"`
	ApiSecret string `json:"api_secret"`
}

func ReadAuthFile(authPath string) (AuthConfig, error) {

	var auth AuthConfig
	err := UnmarshalJsonFile(authPath, &auth)
	if err != nil {
		return AuthConfig{}, err
	}

	err = validateAuthFields(auth)
	if err != nil {
		return AuthConfig{}, err
	}

	return auth, nil
}

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

func validateAuthFields(auth AuthConfig) error {

	if auth.ApiKey == "" {
		return errors.New("Auth.ApiKey is empty")
	}

	if auth.ApiSecret == "" {
		return errors.New("Auth.ApiSecret is empty")
	}

	return nil
}
