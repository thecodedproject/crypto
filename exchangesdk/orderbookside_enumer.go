// Code generated by "enumer -type=OrderBookSide -trimprefix=OrderBookSide -json -text -transform=snake"; DO NOT EDIT.

//
package exchangesdk

import (
	"encoding/json"
	"fmt"
)

const _OrderBookSideName = "unknownbidasksentinal"

var _OrderBookSideIndex = [...]uint8{0, 7, 10, 13, 21}

func (i OrderBookSide) String() string {
	if i < 0 || i >= OrderBookSide(len(_OrderBookSideIndex)-1) {
		return fmt.Sprintf("OrderBookSide(%d)", i)
	}
	return _OrderBookSideName[_OrderBookSideIndex[i]:_OrderBookSideIndex[i+1]]
}

var _OrderBookSideValues = []OrderBookSide{0, 1, 2, 3}

var _OrderBookSideNameToValueMap = map[string]OrderBookSide{
	_OrderBookSideName[0:7]:   0,
	_OrderBookSideName[7:10]:  1,
	_OrderBookSideName[10:13]: 2,
	_OrderBookSideName[13:21]: 3,
}

// OrderBookSideString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func OrderBookSideString(s string) (OrderBookSide, error) {
	if val, ok := _OrderBookSideNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to OrderBookSide values", s)
}

// OrderBookSideValues returns all values of the enum
func OrderBookSideValues() []OrderBookSide {
	return _OrderBookSideValues
}

// IsAOrderBookSide returns "true" if the value is listed in the enum definition. "false" otherwise
func (i OrderBookSide) IsAOrderBookSide() bool {
	for _, v := range _OrderBookSideValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for OrderBookSide
func (i OrderBookSide) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for OrderBookSide
func (i *OrderBookSide) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("OrderBookSide should be a string, got %s", data)
	}

	var err error
	*i, err = OrderBookSideString(s)
	return err
}

// MarshalText implements the encoding.TextMarshaler interface for OrderBookSide
func (i OrderBookSide) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for OrderBookSide
func (i *OrderBookSide) UnmarshalText(text []byte) error {
	var err error
	*i, err = OrderBookSideString(string(text))
	return err
}
