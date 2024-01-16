//go:generate mockery --name OauthRest  --inpackage --case underscore
//go:generate mockery --name OauthService --inpackage --case underscore
package authz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/slausonio/siocore"
)

var ctx = context.Background()

type TokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

func (tr *TokenResp) MarshalBinary() ([]byte, error) {
	return json.Marshal(tr)
}

func (tr *TokenResp) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &tr); err != nil {
		return err
	}

	return nil
}

type IntrospectResp struct {
	Active     bool   `json:"active"`
	Scope      string `json:"scope"`
	ClientID   string `json:"client_id"`
	Expiration int64  `json:"exp"`
	IssuedAt   int64  `json:"iat"`
	Nbf        int64  `json:"nbf"`
	Issuer     string `json:"iss"`
	TokenType  string `json:"type"`
	TokenUse   string `json:"token_use"`
}

func (ir IntrospectResp) MarshalBinary() ([]byte, error) {
	return json.Marshal(ir)
}

func (ir *IntrospectResp) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &ir); err != nil {
		return fmt.Errorf("error unmarshalling introspect response: %w", err)
	}

	return nil
}

type OauthClient struct {
	clientID     string
	clientSecret string
	issuerBase   string
	adminBase    string
	restUtils    HttpHelpers
	log          slog.Logger
}

type OauthRest interface {
	Create() (*TokenResp, error)
	Introspect(authHeader string) (*IntrospectResp, error)
}

func NewOauthClient(log *slog.Logger, appEnv siocore.Env) *OauthClient {
	authzEnvKeys := []string{EnvKeyOauthClientID, EnvKeyOauthClientSecret, EnvKeyOauthAdminBase, EnvKeyOauthIssuerBase}

	checkAuthzEnv(authzEnvKeys, appEnv)

	return &OauthClient{
		clientID:     appEnv[EnvKeyOauthClientID],
		clientSecret: appEnv[EnvKeyOauthClientSecret],
		issuerBase:   appEnv[EnvKeyOauthIssuerBase],
		adminBase:    appEnv[EnvKeyOauthAdminBase],
		restUtils:    NewRestHelpers(),
		log:          log,
	}
}

func (c *OauthClient) Create() (*TokenResp, error) {
	params := url.Values{}
	params.Add("client_id", c.clientID)
	params.Add("client_secret", c.clientSecret)
	params.Add("grant_type", `client_credentials`)
	params.Add("scope", `write`)
	body := strings.NewReader(params.Encode())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.issuerBase+"/oauth2/token", body)
	if err != nil {
		return nil, fmt.Errorf("error creating oauth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tr := new(TokenResp)
	err = c.restUtils.DoHttpRequestAndParse(req, tr)
	if err != nil {
		return nil, fmt.Errorf("error doing oauth request: %w", err)
	}
	return tr, nil
}

func (c *OauthClient) Introspect(authHeader string) (*IntrospectResp, error) {
	introspectStart := time.Now()
	data := url.Values{
		"token": {authHeader},
	}

	requestURL := c.adminBase + "/admin/oauth2/introspect"
	introspectResp := new(IntrospectResp)
	err := c.restUtils.PostHttpFormAndParse(requestURL, data, introspectResp)
	if err != nil {
		return nil, fmt.Errorf("error doing introspect request: %w", err)
	}

	c.log.Info("introspect token http call took", time.Since(introspectStart))

	return introspectResp, nil
}

const (
	authCacheKey  = "oauth_token/"
	introspectKey = "introspect/"
)

type OauthSvc struct {
	redis *redis.Client
	oauth OauthRest
	appID string
}

type OauthService interface {
	Get() (*TokenResp, error)
	Introspect(token string) (*IntrospectResp, error)
}

func NewOauthService(log *slog.Logger, redis *redis.Client, appEnv siocore.Env) *OauthSvc {
	return &OauthSvc{
		redis: redis,
		oauth: NewOauthClient(log, appEnv),
		appID: appEnv.Value(siocore.EnvKeyAppName),
	}
}

func (s *OauthSvc) Get() (*TokenResp, error) {
	cachedAuth := s.redis.Get(ctx, authCacheKey+s.appID)
	if cachedAuth.Err() != nil {
		if errors.Is(cachedAuth.Err(), redis.Nil) {
			return s.createOauth()
		} else {
			return nil, cachedAuth.Err()
		}
	}

	if cachedAuth.Val() == "" {
		return s.createOauth()
	}

	tkResp := &TokenResp{}

	err := json.Unmarshal([]byte(cachedAuth.Val()), tkResp)
	if err != nil {
		return nil, err
	}

	return tkResp, nil
}

func (s *OauthSvc) Introspect(token string) (*IntrospectResp, error) {
	introspectStart := time.Now()
	cachedIntro := s.redis.Get(ctx, introspectKey+token)

	if cachedIntro.Err() != nil {
		if errors.Is(cachedIntro.Err(), redis.Nil) {
			return s.handleIntrospect(token)
		} else {
			return nil, cachedIntro.Err()
		}
	} else if cachedIntro.Val() == "" {
		return s.handleIntrospect(token)
	}

	logrus.Debug("Token came from cache")

	introspectResp := &IntrospectResp{}
	err := json.Unmarshal([]byte(cachedIntro.Val()), introspectResp)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Introspect took %s", time.Since(introspectStart))

	return introspectResp, nil
}

func (s *OauthSvc) createOauth() (*TokenResp, error) {
	tkResp, err := s.oauth.Create()
	if err != nil {
		return nil, err
	}

	setResult := s.redis.Set(
		ctx,
		authCacheKey+s.appID,
		tkResp,
		time.Second*time.Duration(tkResp.ExpiresIn),
	)
	if setResult.Err() != nil {
		return nil, setResult.Err()
	}

	return tkResp, nil
}

func (s *OauthSvc) handleIntrospect(token string) (*IntrospectResp, error) {
	introspectResp, err := s.oauth.Introspect(token)
	if err != nil {
		return nil, err
	}

	expiration := introspectResp.Expiration
	cacheExp := expiration - time.Now().Unix()
	sc := s.redis.Set(ctx, introspectKey+token, introspectResp, time.Second*time.Duration(cacheExp))
	if sc.Err() != nil {
		return nil, sc.Err()
	}

	return introspectResp, nil
}
