package crypto

import (
	"encoding/json"
	"fmt"
	"strings"
)

const seperator = "__"

func (e Exchange) String() string {

	return e.Provider.String() + seperator + e.Pair.String()
}

func ExchangeString(s string) (Exchange, error) {

	split := strings.Split(s, seperator)

	if len(split) != 2 {
		return Exchange{}, fmt.Errorf("%s is not a valid crypto.Exchange; got %d elements when split, but expected 2", s, len(split))
	}

	apiProvider, err := ApiProviderString(split[0])
	if err != nil {
		return Exchange{}, fmt.Errorf(
			"%s is not a valid crypto.Echange: %w",
			s,
			err,
		)
	}

	pair, err := PairString(split[1])
	if err != nil {
		return Exchange{}, fmt.Errorf(
			"%s is not a valid crypto.Echange: %w",
			s,
			err,
		)
	}

	return Exchange{
		Provider: apiProvider,
		Pair: pair,
	}, nil
}

func (e Exchange) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *Exchange) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("Exchange should be a string, got %s", data)
	}

	var err error
	*e, err = ExchangeString(s)
	return err
}

func (e Exchange) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

func (e *Exchange) UnmarshalText(text []byte) error {
	var err error
	*e, err = ExchangeString(string(text))
	return err
}
