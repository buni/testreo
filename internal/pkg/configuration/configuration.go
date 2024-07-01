package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/spf13/viper"
)

func NewConfiguration() (*Configuration, error) {
	config := &Configuration{}

	method := reflect.ValueOf(config).MethodByName("SetDefaults")
	if method.IsValid() {
		method.Call(nil)
	}

	viper.AutomaticEnv()
	viper.SetConfigType(`json`)

	jsonConf, err := json.Marshal(config)
	if err != nil {
		return config, fmt.Errorf("failed to marshal config to json: %w", err)
	}

	err = viper.MergeConfig(bytes.NewBuffer(jsonConf))
	if err != nil {
		return config, fmt.Errorf("failed to merge config: %w", err)
	}

	err = viper.Unmarshal(config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}
