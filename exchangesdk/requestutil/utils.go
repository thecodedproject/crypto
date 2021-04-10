package requestutil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func FullPath(baseUrl string, paths ...string) *url.URL {

	base, err := url.Parse(baseUrl)
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range paths {
		pUrl, err := url.Parse(p)
		if err != nil {
			log.Fatal(err)
		}

		base = base.ResolveReference(pUrl)
	}

	return base
}

func HttpStatusError(res *http.Response, i ...interface{}) error {

	if len(i) > 0 {
		return fmt.Errorf("https status %d: %s (%s)",
			res.StatusCode,
			res.Status,
			fmt.Sprint(i...),
		)
	}

	return fmt.Errorf("https status %d: %s",
		res.StatusCode,
		res.Status,
	)
}

type RoundTripFunc func(*http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {

	return f(req), nil
}

func ResBodyFromJsonf(
	_ *testing.T,
	jsonStringf string,
	i ...interface{},
) io.ReadCloser {

	jsonString := fmt.Sprintf(jsonStringf, i...)
	jsonBuffer := bytes.NewBufferString(jsonString)
	return ioutil.NopCloser(jsonBuffer)
}

func GetReqBodyValues(
	t *testing.T,
	req *http.Request,
) url.Values {

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	require.NoError(t, err)

	reqValues, err := url.ParseQuery(string(body))
	require.NoError(t, err)

	return reqValues
}
