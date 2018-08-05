package store

import (
	"fmt"
	"github.com/spf13/viper"
)

func readConfig(keys []string) (map[string]string, error) {
	_map := make(map[string]string)
	for _, key := range keys {
		value := viper.GetString(key)
		if value == "" {
			return nil, fmt.Errorf("missing config key: %s", key)
		} else {
			_map[key] = value
		}
	}
	return _map, nil
}
