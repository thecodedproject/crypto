package io

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type AuthConfig struct {
	ApiKey string `json:"api_key"`
	ApiSecret string `json:"api_secret"`
}

type ApiKeysConfig struct {
	ApiKeys map[string]AuthConfig `json:"keys"`
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

func GetApiAuthByName(
	authFilePath string,
	keyName string,
) (AuthConfig, error) {

	var keys ApiKeysConfig
	err := UnmarshalJsonFile(authFilePath, &keys)
	if err != nil {
		return AuthConfig{}, err
	}

	if len(keys.ApiKeys) == 0 {
		return AuthConfig{}, errors.New("Auth.Keys is empty - no API keys found")
	}

	for name, val := range keys.ApiKeys {
		err = validateAuthFields(val)
		if err != nil {
			return AuthConfig{}, err
		}

		if name == keyName {
			return val, nil
		}
	}

	return AuthConfig{}, fmt.Errorf("No such api key `%s` in Auth.Keys", keyName)
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
