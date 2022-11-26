package cmd

import (
	"create-cli/internal/sonarqube"

	"github.com/spf13/cobra"
)

func SonarQube(args []string) {
	sonarqube.SonarQube()
}

func init() {
	rootCmd.AddCommand(sonarqubeCmd)
}

var sonarqubeCmd = &cobra.Command{
	Use:   "sonarqube",
	Short: "Configures SonarQube",
	Run: func(cmd *cobra.Command, args []string) {
		SonarQube(args)
	},
}
