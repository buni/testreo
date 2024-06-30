package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/spf13/viper"
)

func NewConfiguration[T any]() (conf *T, err error) {
	conf = new(T)

	method := reflect.ValueOf(conf).MethodByName("SetDefaults")
	if method.IsValid() {
		method.Call(nil)
	}

	viper.AutomaticEnv()
	viper.SetConfigType(`json`)

	jsonConf, err := json.Marshal(conf)
	if err != nil {
		return conf, fmt.Errorf("failed to : %w", err)
	}

	err = viper.MergeConfig(bytes.NewBuffer(jsonConf))
	if err != nil {
		return conf, fmt.Errorf("failed to : %w", err)
	}

	err = viper.Unmarshal(conf)
	if err != nil {
		return conf, fmt.Errorf("failed to : %w", err)
	}

	return conf, nil
}
