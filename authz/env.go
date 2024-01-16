package authz

import "github.com/slausonio/siocore"
import "fmt"

const (
	EnvKeyOauthClientID     = "OAUTH_CLIENT_ID"
	EnvKeyOauthClientSecret = "OAUTH_CLIENT_SECRET"
	EnvKeyOauthIssuerBase   = "OAUTH_ISSUER_BASE"
	EnvKeyOauthAdminBase    = "OAUTH_ADMIN_BASE"
)

// checkAuthzEnv validates required authz variables are present.  If not, the thread will panic
func checkAuthzEnv(oauthEnvVarKeys []string, env siocore.Env) {
	for _, key := range oauthEnvVarKeys {
		value, present := env.LookupValue(key)
		if !present || value == "" {
			panic(
				fmt.Sprintf(
					"The environment variable %s is not present.\n Unable to start application.",
					key,
				),
			)
		}
	}
}
