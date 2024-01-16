package authz

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

const (
	testTokenKey       = "oauth_token/test-app"
	introspectBuiltKey = "introspect/test-token"
)

var (
	hourFromNow             = time.Now().Add(time.Hour).Unix()
	now                     = time.Now().Unix()
	expectedIntroExpiration = hourFromNow - now
	expectedToken           = &TokenResp{
		AccessToken: "test-token",
		ExpiresIn:   5999,
		Scope:       "read",
		TokenType:   "bearer",
	}
	expectedIntrospect = &IntrospectResp{
		Active:     true,
		Scope:      "read",
		ClientID:   "test-client",
		Expiration: hourFromNow,
		IssuedAt:   now,
		Nbf:        now,
		Issuer:     "test-issuer",
		TokenType:  "access_token",
	}
)

func TestMockOauthRest_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockHttpHandler := NewMockHttpHelpers(t)

		tc := &OauthClient{
			restUtils: mockHttpHandler,
		}

		err := os.Setenv("OAUTH_CLIENT_ID", "test_client_id")
		if err != nil {
			t.Fatal(err)
		}
		err = os.Setenv("OAUTH_CLIENT_SECRET", "test_client_secret")
		if err != nil {
			t.Fatal(err)
		}

		ti := new(TokenResp)

		mockHttpHandler.On("DoHttpRequestAndParse", mock.AnythingOfType("*http.Request"), ti).
			Return(nil)

		result, err := tc.Create()
		if err != nil {
			t.Errorf("TestCreateToken() error = %v", err)
			return
		}

		if result == nil {
			t.Errorf("CreateToken() did not set AccessToken")
			return
		}

		err = os.Unsetenv("OAUTH_CLIENT_ID")
		if err != nil {
			t.Fatal(err)
		}
		err = os.Unsetenv("OAUTH_CLIENT_SECRET")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error happened", func(t *testing.T) {
		mockHttpHandler := NewMockHttpHelpers(t)

		tc := &OauthClient{
			restUtils: mockHttpHandler,
		}

		err := os.Setenv("OAUTH_CLIENT_ID", "test_client_id")
		if err != nil {
			t.Fatal(err)
		}
		err = os.Setenv("OAUTH_CLIENT_SECRET", "test_client_secret")
		if err != nil {
			t.Fatal(err)
		}

		ti := new(TokenResp)

		mockHttpHandler.On("DoHttpRequestAndParse", mock.AnythingOfType("*http.Request"), ti).
			Return(fmt.Errorf("biohazard"))

		_, err = tc.Create()
		if err == nil {
			t.Error("Error not returned")
			return
		}
	})
}

func TestMockOauthRest_Introspect(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockHttpHandler := NewMockHttpHelpers(t)

		tc := &OauthClient{
			restUtils: mockHttpHandler,
		}

		err := os.Setenv("OAUTH_CLIENT_ID", "test_client_id")
		if err != nil {
			t.Fatal(err)
		}
		err = os.Setenv("OAUTH_CLIENT_SECRET", "test_client_secret")
		if err != nil {
			t.Fatal(err)
		}

		ir := new(IntrospectResp)

		mockHttpHandler.On("PostHttpFormAndParse", mock.AnythingOfType("string"), mock.AnythingOfType("url.Values"), ir).
			Return(nil)

		result, err := tc.Introspect("sdf")
		if err != nil {
			t.Errorf("TestIntrospect error = %v", err)
			return
		}

		if result == nil {
			t.Errorf("Introspect did not set AccessToken")
			return
		}

		err = os.Unsetenv("OAUTH_CLIENT_ID")
		if err != nil {
			t.Fatal(err)
		}
		err = os.Unsetenv("OAUTH_CLIENT_SECRET")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error happened", func(t *testing.T) {
		mockHttpHandler := NewMockHttpHelpers(t)

		tc := &OauthClient{
			restUtils: mockHttpHandler,
		}

		err := os.Setenv("OAUTH_CLIENT_ID", "test_client_id")
		if err != nil {
			t.Fatal(err)
		}
		err = os.Setenv("OAUTH_CLIENT_SECRET", "test_client_secret")
		if err != nil {
			t.Fatal(err)
		}

		ir := new(IntrospectResp)

		mockHttpHandler.On("PostHttpFormAndParse", mock.AnythingOfType("string"), mock.AnythingOfType("url.Values"), ir).
			Return(fmt.Errorf("biohazard"))

		_, err = tc.Introspect("sdf")
		if err == nil {
			t.Error("Error not returned")
			return
		}
	})
}

func setUp(t *testing.T) (*OauthSvc, *MockOauthRest, redismock.ClientMock) {
	t.Helper()
	db, mRc := redismock.NewClientMock()
	mClient := NewMockOauthRest(t)
	svc := &OauthSvc{
		redis: db,
		oauth: mClient,
		appID: "test-app",
	}
	return svc, mClient, mRc
}

func TestOauthSvc_Get(t *testing.T) {
	t.Run("token in redis", func(t *testing.T) {
		svc, _, mRc := setUp(t)

		tokenJson, err := json.Marshal(expectedToken)
		if err != nil {
			t.Fatal(err)
		}

		mRc.ExpectGet(testTokenKey).SetVal(string(tokenJson))

		token, err := svc.Get()
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, token)
	})

	t.Run("Error Unmarshalling token", func(t *testing.T) {
		svc, _, mRc := setUp(t)
		mRc.ExpectGet(testTokenKey).SetVal(`{est": "test"}`)

		token, err := svc.Get()
		assert.Error(t, err)
		assert.Nilf(t, token, "token should be nil")
	})

	t.Run("error getting token from redis", func(t *testing.T) {
		svc, _, mRc := setUp(t)
		mRc.ExpectGet(testTokenKey).SetErr(errors.New("test error"))
		_, err := svc.Get()
		assert.Error(t, err)
	})

	t.Run("nil token in redis,  successful creation", func(t *testing.T) {
		svc, mOr, mRc := setUp(t)
		mRc.ExpectGet(testTokenKey).RedisNil()
		mOr.On("Create").
			Return(expectedToken, nil)
		mRc.ExpectSet(testTokenKey, expectedToken, time.Second*time.Duration(expectedToken.ExpiresIn)).
			SetVal("OK")
		token, err := svc.Get()
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, token)
	})

	t.Run("nil token in redis bad creation", func(t *testing.T) {
		svc, mOr, mRc := setUp(t)
		mRc.ExpectGet(testTokenKey).RedisNil()
		mOr.On("Create").
			Return(nil, errors.New("test error"))
		token, err := svc.Get()
		assert.Error(t, err)
		assert.Nil(t, token)
	})

	t.Run("no token in redis, successful creation", func(t *testing.T) {
		svc, mOr, mRc := setUp(t)

		mRc.ExpectGet(testTokenKey).SetVal("")

		mOr.On("Create").
			Return(expectedToken, nil)
		mRc.ExpectSet(testTokenKey, expectedToken, time.Second*time.Duration(expectedToken.ExpiresIn)).
			SetVal("OK")

		token, err := svc.Get()
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, token)
	})

	t.Run("no token in redis, bad creation", func(t *testing.T) {
		svc, mOr, mRc := setUp(t)
		mRc.ExpectGet(testTokenKey).SetVal("")

		mOr.On("Create").
			Return(nil, errors.New("test error"))

		token, err := svc.Get()
		assert.Error(t, err)
		assert.Nil(t, token)
	})

	t.Run("nil token in redis, successful creation, fail on caching", func(t *testing.T) {
		svc, mOr, mRc := setUp(t)
		mRc.ExpectGet(testTokenKey).RedisNil()

		mOr.On("Create").
			Return(expectedToken, nil)
		mRc.ExpectSet(testTokenKey, expectedToken, time.Second*time.Duration(expectedToken.ExpiresIn)).
			SetErr(errors.New("test error"))

		token, err := svc.Get()
		assert.Error(t, err)
		assert.Nil(t, token)
	})

	t.Run("no token in redis, successful creation, fail on caching", func(t *testing.T) {
		svc, mOr, mRc := setUp(t)
		mRc.ExpectGet(testTokenKey).SetVal("")

		mOr.On("Create").
			Return(expectedToken, nil)
		mRc.ExpectSet(testTokenKey, expectedToken, time.Second*time.Duration(expectedToken.ExpiresIn)).
			SetErr(errors.New("test error"))

		token, err := svc.Get()
		assert.Error(t, err)
		assert.Nil(t, token)
	})
}

func TestOauthSvc_Introspect(t *testing.T) {
	t.Run("introspect in redis", func(t *testing.T) {
		svc, _, mRc := setUp(t)

		introspectJson, err := json.Marshal(expectedIntrospect)
		if err != nil {
			t.Fatal(err)
		}

		mRc.ExpectGet(introspectBuiltKey).SetVal(string(introspectJson))

		introspect, err := svc.Introspect("test-token")
		assert.NoError(t, err)
		assert.Equal(t, expectedIntrospect, introspect)
	})

	t.Run("Error Unmarshalling introspect", func(t *testing.T) {
		svc, _, mRc := setUp(t)
		mRc.ExpectGet(introspectKey).SetVal(`{est": "test"}`)

		introspect, err := svc.Introspect("test-token")
		assert.Error(t, err)
		assert.Nilf(t, introspect, "introspect should be nil")
	})

	t.Run("error getting introspect from redis", func(t *testing.T) {
		svc, _, mRc := setUp(t)
		mRc.ExpectGet(introspectBuiltKey).SetErr(errors.New("test error"))
		_, err := svc.Introspect("test-token")
		assert.Error(t, err)
	})

	t.Run("nilIntrospect_successfulCreation", func(t *testing.T) {
		svc, mOr, mRc := setUp(t)

		mRc.ExpectGet(introspectBuiltKey).RedisNil()

		mOr.On("Introspect", "test-token").
			Return(expectedIntrospect, nil)
		mRc.ExpectSet(introspectBuiltKey, expectedIntrospect, time.Second*time.Duration(expectedIntroExpiration)).
			SetVal("OK")

		introspect, err := svc.Introspect("test-token")
		assert.NoError(t, err)
		assert.Equal(t, expectedIntrospect, introspect)
	})

	t.Run("nilIntrospect_badCreation", func(t *testing.T) {
		svc, mOr, mRc := setUp(t)
		mRc.ExpectGet(introspectBuiltKey).RedisNil()

		mOr.On("Introspect", "test-token").
			Return(nil, errors.New("test error"))

		introspect, err := svc.Introspect("test-token")
		assert.Error(t, err)
		assert.Nil(t, introspect)
	})

	t.Run("noIntrospect_successfulCreation", func(t *testing.T) {
		svc, mOr, mRc := setUp(t)

		mRc.ExpectGet(introspectBuiltKey).SetVal("")

		mOr.On("Introspect", "test-token").
			Return(expectedIntrospect, nil)
		mRc.ExpectSet(introspectBuiltKey, expectedIntrospect, time.Second*time.Duration(expectedIntroExpiration)).
			SetVal("OK")

		introspect, err := svc.Introspect("test-token")
		assert.NoError(t, err)
		assert.Equal(t, expectedIntrospect, introspect)
	})

	t.Run("noIntrospect_badCreation", func(t *testing.T) {
		svc, mOr, mRc := setUp(t)
		mRc.ExpectGet(introspectBuiltKey).SetVal("")

		mOr.On("Introspect", "test-token").
			Return(nil, errors.New("test error"))

		introspect, err := svc.Introspect("test-token")
		assert.Error(t, err)
		assert.Nil(t, introspect)
	})

	t.Run("no introspect in redis, successful creation, fail on caching", func(t *testing.T) {
		svc, mOr, mRc := setUp(t)
		mRc.ExpectGet(introspectBuiltKey).SetVal("")

		mOr.On("Introspect", "test-token").
			Return(expectedIntrospect, nil)
		mRc.ExpectSet(introspectBuiltKey, expectedIntrospect, time.Second*time.Duration(expectedIntrospect.Expiration)).
			SetErr(errors.New("test error"))

		token, err := svc.Introspect("test-token")
		assert.Error(t, err)
		assert.Nil(t, token)
	})
}
