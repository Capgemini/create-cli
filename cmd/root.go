package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "create-cli",
	Short: "create-cli - a CLI to manage configuration of CREATE",
	Long:  `create-cli can be used to make configuration of CREATE infrastructure and applications easy.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

type k8sFlags struct {
	outCluster bool
}

var createUrl string
var personalAccessToken string

var k8sArgs = &k8sFlags{}

func Execute() {
	rootCmd.PersistentFlags().BoolVarP(&k8sArgs.outCluster, "out-cluster", "", k8sArgs.outCluster, "specifies if the in cluster k8s client has to be used")
	viper.BindPFlag("out-cluster", rootCmd.PersistentFlags().Lookup("out-cluster"))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
