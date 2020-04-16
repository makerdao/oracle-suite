package query

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

// Default retry amount
const defaultRetry = 5

// HTTPRequest default HTTP Request structure
type HTTPRequest struct {
	URL     string
	Headers map[string]string
	Retry   int
}

// HTTPResponse default query engine response
type HTTPResponse struct {
	Body  []byte
	Error error
}

// MakeGetRequest makes GET HTTP request to given `url` with `headers` and in case of error
// it will retry request `retry` amount of times. And only after it (if it's still error) error will be returned.
// Automatically timeout between requests will be calculated using `random`.
// Note for `timeout` waiting this function uses `time.Sleep()` so it will block execution flow.
// Better to be used in go-routine.
func MakeGetRequest(r *HTTPRequest) *HTTPResponse {
	if r == nil {
		return &HTTPResponse{
			Error: fmt.Errorf("failed to make HTTP request to `nil`"),
		}
	}

	// Check for non set Retry
	if r.Retry == 0 {
		r.Retry = defaultRetry
	}

	step := 1
	var res []byte
	var err error

	for step <= r.Retry {
		res, err = doMakeGetRequest(r)
		if err != nil {
			time.Sleep(getTimeout())
			step++
			continue
		}
		// All ok no `err` received
		break
	}

	return &HTTPResponse{
		Body:  res,
		Error: err,
	}
}

func doMakeGetRequest(r *HTTPRequest) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("failed to make HTTP request to `nil`")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", r.URL, nil)
	if err != nil {
		return nil, err
	}
	if r.Headers != nil {
		for k, v := range r.Headers {
			req.Header.Add(k, v)
		}
	}
	// Perform HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("failed to make HTTP request to %s, got %d status code", r.URL, resp.StatusCode)
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// getTimeout geenrate random timeout between retry queries
func getTimeout() time.Duration {
	return time.Duration(100 + rand.Intn(900))
}
