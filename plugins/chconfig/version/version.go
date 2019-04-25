package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

const (
	programName = "chconfig"
	version     = "1.0.0"
)

// Cmd returns the Cobra Command for Version
func Cmd() *cobra.Command {
	return versionCmd
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print chconfig version",
	Long:  "Print current version of the chconfig",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		fmt.Print(versionInfo())
		return nil
	},
}

func versionInfo() string {
	return fmt.Sprintf("%s:\n Version: %s\n Go version: %s\n OS/Arch: %s\n",
		programName,
		version,
		runtime.Version(),
		fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
}
