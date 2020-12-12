package io

import (
	"errors"
	"fmt"
	"github.com/thecodedproject/crypto"
)


func ReadAuthFile(authPath string) (map[string]crypto.AuthConfig, error) {

	var keys struct {
		ApiKeys map[string]crypto.AuthConfig `json:"keys"`
	}
	err := UnmarshalJsonFile(authPath, &keys)
	if err != nil {
		return nil, err
	}

	for authName, auth := range keys.ApiKeys {
		err = validateAuthFields(authName, auth)
		if err != nil {
			return nil, err
		}
	}

	return keys.ApiKeys, nil
}

func GetApiAuthByName(
	authFilePath string,
	keyName string,
) (crypto.AuthConfig, error) {

	var keys struct {
		ApiKeys map[string]crypto.AuthConfig `json:"keys"`
	}
	err := UnmarshalJsonFile(authFilePath, &keys)
	if err != nil {
		return crypto.AuthConfig{}, err
	}

	if len(keys.ApiKeys) == 0 {
		return crypto.AuthConfig{}, errors.New("Auth.Keys is empty - no API keys found")
	}

	for name, val := range keys.ApiKeys {
		err = validateAuthFields(name, val)
		if err != nil {
			return crypto.AuthConfig{}, err
		}

		if name == keyName {
			return val, nil
		}
	}

	return crypto.AuthConfig{}, fmt.Errorf("No such api key `%s` in Auth.Keys", keyName)
}

func validateAuthFields(authName string, auth crypto.AuthConfig) error {

	if auth.ApiExchange == crypto.ExchangeUnknown {
		return fmt.Errorf("Auth[%s].Exchange is not a known exchange", authName)
	}

	if auth.ApiKey == "" {
		return fmt.Errorf("Auth[%s].ApiKey is empty", authName)
	}

	if auth.ApiSecret == "" {
		return fmt.Errorf("Auth[%s].ApiSecret is empty", authName)
	}

	return nil
}

