package query

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

// MakeGetRequest makes GET HTTP request to given URL
func MakeGetRequest(url string) ([]byte, error) {
	return MakeGetRequestWithHeaders(url, nil)
}

// MakeGetRequestWithHeaders makes GET HTTP request to given URL
func MakeGetRequestWithHeaders(url string, headers map[string]string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	// Perform HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("failed to make HTTP request to %s, got %d status code", url, resp.StatusCode)
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// MakeGetRequestWithRetry makes GET HTTP request to given `url` with `headers` and in case of error
// it will retry request `retry` amount of times. And only after it (if it's still error) error will be returned.
// Automatically timeout between requests will be calculated using `random`.
// Note for `timeout` waiting this function uses `time.Sleep()` so it will block execution flow.
// Better to be used in go-routine.
func MakeGetRequestWithRetry(url string, headers map[string]string, retry int) ([]byte, error) {
	step := 1
	var res []byte
	var err error

	for step <= retry {
		res, err = MakeGetRequestWithHeaders(url, headers)
		if err != nil {
			time.Sleep(getTimeout())
			step++
			continue
		}
		// All ok no `err` received
		break
	}

	return res, err
}

// getTimeout geenrate random timeout between retry queries
func getTimeout() time.Duration {
	return time.Duration(100 + rand.Intn(900))
}
