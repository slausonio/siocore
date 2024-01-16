package siogo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockRoundTripper struct {
	expectedReq *http.Request
	response    *http.Response
	err         error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req != m.expectedReq {
		return nil, fmt.Errorf("unexpected request: %+v", req)
	}
	return m.response, m.err
}

func TestDoHttpRequest(t *testing.T) {
	// Create a mock HTTP server to return a response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()

	// Create a request to the mock server
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mocked http.RoundTripper that returns the expected response
	expectedResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"status":"ok"}`)),
	}
	rt := &mockRoundTripper{expectedReq: req, response: expectedResponse, err: nil}

	// Create an http.Client that uses the mocked RoundTripper
	client := &http.Client{Transport: rt}

	// Create a RestHelpers instance that uses the mocked http.Client
	restHelpers := &RestHelpers{client: client}

	// Call the method being tested
	resp, err := restHelpers.DoHttpRequest(req)
	// Check that the expected response was returned
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != expectedResponse.StatusCode {
		t.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	expectedRespBody, err := io.ReadAll(expectedResponse.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(respBody, expectedRespBody) {
		t.Errorf("unexpected response body: %s", respBody)
	}
}

func TestDoHttpRequestAndParse(t *testing.T) {
	// Create a mock HTTP server to return a response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()

	// Create a request to the mock server
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mocked http.RoundTripper that returns the expected response
	expectedResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"status":"ok"}`)),
	}
	rt := &mockRoundTripper{expectedReq: req, response: expectedResponse, err: nil}

	// Create an http.Client that uses the mocked RoundTripper
	client := &http.Client{Transport: rt}

	// Create a RestHelpers instance that uses the mocked http.Client
	restHelpers := &RestHelpers{client: client}

	// Call the method being tested
	var v struct{ Status string }
	err = restHelpers.DoHttpRequestAndParse(req, &v)

	// Check that the expected response was returned
	if err != nil {
		t.Fatal(err)
	}
	if v.Status != "ok" {
		t.Errorf("unexpected response: %+v", v)
	}

	// Call the method being tested with an unexpected request
	req2, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = restHelpers.DoHttpRequestAndParse(req2, &v)
	if err == nil {
		t.Error("expected an error")
	}
}

func TestBuildRequest(t *testing.T) {
	method := "GET"
	url := "http://example.com"
	accessToken := "testToken"

	t.Run("With Body", func(t *testing.T) {
		bodyReader := strings.NewReader(`{"key": "value"}`)

		req, err := BuildRequest(method, url, bodyReader, accessToken)
		assert.NoError(t, err)
		assert.NotNil(t, req)

		assert.Equal(t, method, req.Method)
		assert.Equal(t, url, req.URL.String())
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer "+accessToken, req.Header.Get("Authorization"))
	})

	t.Run("nil body", func(t *testing.T) {
		req, err := BuildRequest(method, url, nil, accessToken)
		assert.NoError(t, err)
		assert.NotNil(t, req)

		assert.Equal(t, method, req.Method)
		assert.Equal(t, url, req.URL.String())
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer "+accessToken, req.Header.Get("Authorization"))
	})
}

func TestBuildRequest_Error(t *testing.T) {
	_, err := BuildRequest("", "", nil, "")
	assert.Error(t, err)
}

func TestBuildRequestWithBody(t *testing.T) {
	method := "POST"
	url := "http://example.com"
	accessToken := "testToken"
	reqBody := map[string]string{
		"key": "value",
	}

	req, err := BuildRequestWithBody(method, url, reqBody, accessToken)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	assert.Equal(t, method, req.Method)
	assert.Equal(t, url, req.URL.String())
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer "+accessToken, req.Header.Get("Authorization"))

	expectedBody, _ := json.Marshal(reqBody)
	assert.Equal(t, string(expectedBody), readBody(t, req))
}

func TestBuildRequestWithBody_Error(t *testing.T) {
	_, err := BuildRequestWithBody("POST", "", make(chan int), "")
	assert.Error(t, err)
}

func readBody(t *testing.T, req *http.Request) string {
	t.Helper()

	bodyBytes, _ := io.ReadAll(req.Body)
	return string(bodyBytes)
}
