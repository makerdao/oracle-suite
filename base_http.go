package gofer

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// MakeGetRequest makes GET HTTP request to given URL
func MakeGetRequest(url string) ([]byte, error) {
	return MakeGetRequestWithHeaders(url, nil)
}

// MakeGetRequest makes GET HTTP request to given URL
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
