//go:generate mockery --name HttpHelpers --inpackage --case underscore
package siogo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type RestHelpers struct {
	client *http.Client
}


func NewRestHelpers() *RestHelpers {
	return &RestHelpers{
		client: http.DefaultClient,
	}
}

func (r *RestHelpers) DoHttpRequest(req *http.Request) (*http.Response, error) {
	res, err := r.ExecuteRequest(req)
	if err != nil {
		return nil, err
	}

	return r.HandleResponse(res)
}

func (r *RestHelpers) PostHttpForm(url string, values url.Values) (*http.Response, error) {
	res, err := r.PostForm(url, values)
	if err != nil {
		return nil, err
	}

	return r.HandleResponse(res)
}

func (r *RestHelpers) DoHttpRequestAndParse(req *http.Request, v interface{}) error {
	res, err := r.ExecuteRequest(req)
	if err != nil {
		return err
	}

	res, err = r.HandleResponse(res)
	if err != nil {
		return err
	}
	return r.ParseResponse(res, v)
}

func (r *RestHelpers) PostHttpFormAndParse(url string, values url.Values, v interface{}) error {
	res, err := r.PostForm(url, values)
	if err != nil {
		return err
	}

	res, err = r.HandleResponse(res)
	if err != nil {
		return err
	}
	return r.ParseResponse(res, v)
}

func (r *RestHelpers) ParseResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf(
			"error parsing response body: %w \n url: %s, statusCode: %d",
			err,
			resp.Request.URL,
			resp.StatusCode,
		)
	}

	return nil
}

func (r *RestHelpers) AbortWithError(err error, code int, c *gin.Context) {
	e := c.AbortWithError(code, err)
	if e != nil {
		log.Fatalf("Error aborting with error: %v", e)
	}
}

func (r *RestHelpers) ExecuteRequest(req *http.Request) (*http.Response, error) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *RestHelpers) PostForm(url string, values url.Values) (*http.Response, error) {
	res, err := http.DefaultClient.PostForm(url, values)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *RestHelpers) HandleResponse(resp *http.Response) (*http.Response, error) {
	if resp.StatusCode >= 400 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf(err.Error())
		}

		return nil, NewAppError(string(b), resp.StatusCode)
	}

	return resp, nil
}

func DecryptAndHandle(request interface{}, c *gin.Context) error {
	enc := NewEncryptionUtil()
	err := c.BindJSON(&request)
	if err != nil {
		return err
	}

	err = enc.DecryptInterface(request)
	if err != nil {
		return err
	}

	return nil
}

// BuildRequest creates a new http.Request with the given method, url and bodyReader. Also, adding the required headers.
func BuildRequest(method string, url string, bodyReader *strings.Reader, accessToken string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating http request request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	return req, nil
}

// BuildRequestWithBody marshals an interface and creates http.Request with body.
func BuildRequestWithBody(method string, url string, reqBody any, accessToken string) (*http.Request, error) {
	rJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating http request request: %w", err)
	}

	sr := strings.NewReader(string(rJSON))

	return BuildRequest(method, url, sr, accessToken)
}
