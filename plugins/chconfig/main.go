package main

import (
	"os"
	"strings"

	"github.com/hyperledger/fabric/plugins/chconfig/start"
	"github.com/hyperledger/fabric/plugins/chconfig/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var mainCmd = &cobra.Command{Use: "chconfig"}

func main() {
	viper.SetEnvPrefix("chcfg")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	mainFlags := mainCmd.PersistentFlags()
	mainFlags.String("logging-level", "", "Logging level flag")
	viper.BindPFlag("logging_level", mainFlags.Lookup("logging-level"))
	mainFlags.MarkHidden("logging-level")

	mainCmd.AddCommand(version.Cmd())
	mainCmd.AddCommand(start.Cmd())

	if mainCmd.Execute() != nil {
		os.Exit(-1)
	}
}
