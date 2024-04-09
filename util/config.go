package util

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Environment             string        `mapstructure:"ENVIRONMENT"`
	DBSource                string        `mapstructure:"DB_SOURCE"`
	HTTPServerAddress       string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	RedisAddress            string        `mapstructure:"REDIS_ADDRESS"`
	TokenSymmetricKey       string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AuthorizationTypeBearer string        `mapstructure:"AUTHORIZATION_TYPE_BEARER"`
	AuthorizationPayloadKey string        `mapstructure:"AUTHORIZATION_PAYLOAD_KEY"`
	AuthorizationHeaderKey  string        `mapstructure:"AUTHORIZATION_HEADER_KEY"`
	AccessTokenDuration     time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
}

func LoadConfig(path string) (config Config, err error) {

	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)

	viper.AutomaticEnv()

	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))

		return
	}

	err = viper.Unmarshal(&config)
	return
}
