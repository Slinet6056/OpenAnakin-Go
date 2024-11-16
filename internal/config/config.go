package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Models map[string]int `mapstructure:"models"`
}

var AppConfig Config

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("../..")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	modelsMap := make(map[string]int)
	models := viper.GetStringMap("models")
	for key, value := range models {
		if v, ok := value.(int); ok {
			modelsMap[key] = v
		}
	}

	AppConfig = Config{
		Models: modelsMap,
	}

	return nil
}
