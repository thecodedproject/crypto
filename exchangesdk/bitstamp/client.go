package bitstamp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto/exchangesdk"
	utiltime "github.com/thecodedproject/crypto/util/time"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	httpsPrefix = "https://"
	bitstampDomain = "www.bitstamp.net"
)

var ErrBadCheckSignature = fmt.Errorf("Bad check signature on response")

type client struct {

	apiKey string
	apiSecret string
	httpClient *http.Client
}

func NewClient(apiKey, apiSecret string) (*client, error) {

	return &client{
		apiKey: apiKey,
		apiSecret: apiSecret,
		httpClient: http.DefaultClient,
	}, nil
}

type roundTripFunc func(*http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {

	return f(req), nil
}

func NewClientForTesting(
	t *testing.T,
	apiKey string,
	apiSecret string,
	handler func(req *http.Request) *http.Response,
) *client {

	return &client{
		apiKey: apiKey,
		apiSecret: apiSecret,
		httpClient: &http.Client{
			Transport: roundTripFunc(handler),
		},
	}
}

func (c *client) LatestPrice(ctx context.Context) (decimal.Decimal, error) {

	req, err := http.NewRequest(
		"GET",
		makeFullUrl("/api/v2/ticker/btceur/"),
		nil,
	)
	if err != nil {
		return decimal.Decimal{}, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return decimal.Decimal{}, err
	}

	if res.StatusCode != http.StatusOK {
		return decimal.Decimal{}, httpStatusError(res)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return decimal.Decimal{}, err
	}

	latestPrice := struct{
		Val decimal.Decimal `json:"last"`
	}{}

	err = json.Unmarshal(body, &latestPrice)
	if err != nil {
		return decimal.Decimal{}, err
	}

	return latestPrice.Val, nil
}

func (c *client) PostLimitOrder(ctx context.Context, order exchangesdk.Order) (string, error) {

	var path string
	switch order.Type {
	case exchangesdk.OrderTypeBid:
		path = "/api/v2/buy/btceur/"
	case exchangesdk.OrderTypeAsk:
		path = "/api/v2/sell/btceur/"
	default:
		return "", fmt.Errorf("Unknown order type")
	}

	values := url.Values{}
	values.Add("price", order.Price.String())
	values.Add("amount", order.Volume.String())

	resBody, err := postRequestWithAuth(
		c.httpClient,
		c.apiKey,
		c.apiSecret,
		path,
		values,
	)
	if err != nil {
		return "", err
	}

	resFields := struct{
		Status *string `json:"status"`
		ErrReason string `json:"reason"`

		Id string `json:"id"`
	}{}

	err = json.Unmarshal(resBody, &resFields)
	if err != nil {
		return "", err
	}

	if resFields.Status != nil {
		return "", fmt.Errorf(
			"Error posting limit order (status='%s'): %s",
			*resFields.Status,
			resFields.ErrReason,
		)
	}

	return resFields.Id, nil
}

func (c *client) StopOrder(ctx context.Context, orderId string) error {

	path := "/api/v2/cancel_order/"

	values := url.Values{}
	values.Add("id", orderId)

	resBody, err := postRequestWithAuth(
		c.httpClient,
		c.apiKey,
		c.apiSecret,
		path,
		values,
	)
	if err != nil {
		return err
	}

	resFields := struct{
		Error *string `json:"error"`
	}{}

	err = json.Unmarshal(resBody, &resFields)
	if err != nil {
		return err
	}

	if resFields.Error != nil {
		return fmt.Errorf("%s", *resFields.Error)
	}

	return nil
}

func makeFullUrl(path string) string {

	return fmt.Sprint(httpsPrefix, bitstampDomain, path)
}

func postRequestWithAuth(
	client *http.Client,
	apiKey string,
	apiSecret string,
	path string,
	values url.Values,
) ([]byte, error) {

	fullUrl := makeFullUrl(path)

	// Bitstamp seems to return an authentication error if the
	// request body is empty (not sure why)
	// So add a dummy value if there are no values here
	if values.Encode() == "" {
		values.Add("some", "value")
	}

	payload := values.Encode()

	reqBody := strings.NewReader(payload)
	reqMethod := "POST"
	req, err := http.NewRequest(reqMethod, fullUrl, reqBody)
	if err != nil {
		return nil, err
	}

	authHeader := "BITSTAMP " + apiKey
	timenow := utiltime.Now()
	timestamp := timenow.Round(time.Millisecond).UnixNano() / 1e6
	timestampStr := strconv.FormatInt(timestamp, 10)
	authVersion := "v2"
	nonce := fmt.Sprintf("%036x", timenow.UnixNano())
	contentType := "application/x-www-form-urlencoded"

	msg := fmt.Sprint(
		authHeader,
		reqMethod,
		bitstampDomain,
		path,
		contentType,
		nonce,
		timestampStr,
		authVersion,
		payload,
	)

	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(msg))
	signature := hex.EncodeToString(
		mac.Sum(nil),
	)

	req.Header.Add("X-Auth", authHeader)
	req.Header.Add("X-Auth-Signature", signature)
	req.Header.Add("X-Auth-Nonce", nonce)
	req.Header.Add("X-Auth-Timestamp", timestampStr)
	req.Header.Add("X-Auth-Version", authVersion)
	req.Header.Add("Content-Type", contentType)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, httpStatusError(res)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	checkMsg := fmt.Sprint(
		nonce,
		timestampStr,
		res.Header.Get("Content-Type"),
		string(body),
	)

	checkMac := hmac.New(sha256.New, []byte(apiSecret))
	checkMac.Write([]byte(checkMsg))
	expectedCheckSignature := hex.EncodeToString(
		checkMac.Sum(nil),
	)

	actualCheckSignature := res.Header.Get("X-Server-Auth-Signature")

	if actualCheckSignature != expectedCheckSignature {
		return nil, ErrBadCheckSignature
	}

	return body, nil
}

func httpStatusError(res *http.Response) error {

	return fmt.Errorf("https status %d: %s", res.StatusCode, res.Status)
}
