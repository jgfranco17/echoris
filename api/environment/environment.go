package environment

import (
	"os"
)

const (
	APPLICATION_ENV_LOCAL string = "local"
	APPLICATION_ENV_DEV   string = "dev"
	APPLICATION_ENV_STAGE string = "stage"
	APPLICATION_ENV_PROD  string = "prod"
)

const (
	ENV_KEY_ENVIRONMENT string = "ENVIRONMENT"
	ENV_KEY_VERSION     string = "APP_VERSION"
	ENV_KEY_LOG_LEVEL   string = "LOG_LEVEL"
)

func IsLocalEnvironment() bool {
	return GetApplicationEnv() == APPLICATION_ENV_LOCAL
}

func GetEnvWithDefault(key string, defaultValue string) string {
	value, present := os.LookupEnv(key)
	if present {
		return value
	}
	return defaultValue
}

func GetApplicationEnv() string {
	return GetEnvWithDefault(ENV_KEY_ENVIRONMENT, APPLICATION_ENV_LOCAL)
}
