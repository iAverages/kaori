package config

import "github.com/spf13/viper"

type Config struct {
	Port          string `mapstructure:"PORT"`
	SpofityId     string `mapstructure:"SPOTIFY_ID"`
	SpofitySecret string `mapstructure:"SPOTIFY_SECRET"`
	Debug         bool   `mapstructure:"DEBUG"`
	Hostname      string `mapstructure:"HOSTNAME"`
	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDatabase int    `mapstructure:"REDIS_DATABASE"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// Default values
	viper.SetDefault("Port", "8080")
	viper.SetDefault("Hostname", "http://localhost:8080")
	viper.SetDefault("RedisHost", "localhost")
	viper.SetDefault("RedisPort", "6379")
	viper.SetDefault("RedisPassword", "")
	viper.SetDefault("RedisDatabase", 0)

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
