package configs

import "github.com/spf13/viper"

type Config struct {
	// database configs
	DBHost         string `mapstructure:"POSTGRES_HOST"`
	DBUsername     string `mapstructure:"POSTGRES_USER"`
	DBUserPassword string `mapstructure:"POSTGRES_PASSWORD"`
	DBName         string `mapstructure:"POSTGRES_NAME"`
	DBPort         string `mapstructure:"POSTGRES_PORT"`

	// app configs
	ServerPort   string `mapstructure:"SERVER_PORT"`
	ClientOrigin string `mapstructure:"CLIENT_ORIGIN"`

	// rmq configs
	RMQUrl         string `mapstructure:"RMQ_URL"`
	RMQQueueName   string `mapstructure:"RMQ_QUEUE_NAME"`
	RMQExchangeKey string `mapstructure:"RMQ_EXCHANGE_KEY"`

	// redis configs
	RedisHost string `mapstructure:"REDIS_HOST"`
	RedisPort string `mapstructure:"REDIS_PORT"`
	RedisDb   string `mapstructure:"REDIS_DB"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("app")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
