package configs

import "github.com/spf13/viper"

type Config struct {
	// database config
	DBHost         string `mapstructure:"POSTGRES_HOST"`
	DBUsername     string `mapstructure:"POSTGRES_USER"`
	DBUserPassword string `mapstructure:"POSTGRES_PASSWORD"`
	DBName         string `mapstructure:"POSTGRES_NAME"`
	DBPort         string `mapstructure:"POSTGRES_PORT"`

	// app config
	ServerPort   string `mapstructure:"SERVER_PORT"`
	ClientOrigin string `mapstructure:"CLIENT_ORIGIN"`

	// payment service config
	PaymentClientBaseUrl string `mapstructure:"PAYMENT_CLIENT_BASE_URL"`

	// rmq config
	RMQUrl         string `mapstructure:"RMQ_URL"`
	RMQQueueName   string `mapstructure:"RMQ_QUEUE_NAME"`
	RMQExchangeKey string `mapstructure:"RMQ_EXCHANGE_KEY"`

	// redis config
	RedisHost string `mapstructure:"REDIS_HOST"`
	RedisPort string `mapstructure:"REDIS_PORT"`
	RedisDb   string `mapstructure:"REDIS_DB"`

	// orchestration
	OrchestrationExpirationTimeSeconds int64  `mapstructure:"ORCHESTRATION_EXPIRATION_TIME_SECONDS"`
	OrchestrationMapName               string `mapstructure:"ORCHESTRATION_MAP_NAME"`
	// produce when rollback starts
	RMQExpiredEventQueue string `mapstructure:"RMQ_EXPIRED_EVENT_QUEUE"`
	// consume and execute rollback of an order
	RMQRollbackEventOrderQueue string `mapstructure:"RMQ_ROLLBACK_ORDER_EVENT_QUEUE"`
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
