package cmd

import "github.com/spf13/viper"

func addConfigPaths() {
	viper.AddConfigPath("/etc/yt-mango")
	viper.AddConfigPath("$HOME/.config/.yt-mango")
}
