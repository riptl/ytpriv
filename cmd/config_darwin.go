package cmd

import "github.com/spf13/viper"

func addConfigPaths() {
	viper.AddConfigPath("$HOME/Library/Application Support")
}
