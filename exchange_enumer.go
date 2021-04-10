package crypto

import (
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
		Pair:     pair,
	}, nil
}

func (e Exchange) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

/* TODO: We don't provider a custom unmarshal function here as we want the exchange type to unmarshalled
	as using the json tags on the struct. This means that the Exchange type doesnt round trip in/out of JSON
	which seems nasty - come up with a better way of solving this.
func (e *Exchange) UnmarshalText(text []byte) error {
	var err error
	*e, err = ExchangeString(string(text))
	return err
}
*/
