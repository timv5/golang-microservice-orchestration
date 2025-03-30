package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	RedisHost string `mapstructure:"REDIS_HOST"`
	RedisPort string `mapstructure:"REDIS_PORT"`
	RedisDB   int    `mapstructure:"REDIS_DB"`

	// this queue is consumed by orchestration service, for expired events
	RMQExpiredEventQueue string `mapstructure:"RMQ_EXPIRED_EVENT_QUEUE"`
	RMQUrl               string `mapstructure:"RMQ_URL"`
	RMQExchangeKey       string `mapstructure:"RMQ_EXCHANGE_KEY"`
	// this queue is consumed by related service, for example order-service
	RMQRollbackEventPaymentQueue string `mapstructure:"RMQ_ROLLBACK_PAYMENT_EVENT_QUEUE"`
	RMQRollbackEventOrderQueue   string `mapstructure:"RMQ_ROLLBACK_ORDER_EVENT_QUEUE"`

	OrchestrationExpirationTimeSeconds int64  `mapstructure:"ORCHESTRATION_EXPIRATION_TIME_SECONDS"`
	OrchestrationMapName               string `mapstructure:"ORCHESTRATION_MAP_NAME"`

	RollbackMinThreads              int `mapstructure:"ROLLBACK_MIN_THREADS"`
	RollbackMaxThreads              int `mapstructure:"ROLLBACK_MAX_THREADS"`
	RollbackMaxThreadsKeepAliveTime int `mapstructure:"ROLLBACK_MAX_THREADS_KEEP_ALIVE_TIME"`

	StaleJobSchedulePeriodMilliseconds int64 `mapstructure:"STALE_JOB_SCHEDULE_PERIOD"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
