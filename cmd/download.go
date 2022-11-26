package cmd

import (
	"create-cli/internal/download"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Download(args []string) {
	download.Download()
}

var cloudProvider string

func init() {
	downloadCmd.Flags().StringVarP(&cloudProvider, "cloud-provider", "", "", "The Cloud Provider that CREATE will exist in")
	downloadCmd.MarkFlagRequired("cloud-provider")
	viper.BindPFlag("cloud-provider", downloadCmd.Flags().Lookup("cloud-provider"))

	downloadCmd.Flags().StringVarP(&personalAccessToken, "pat", "", "", "Personal Access Token for the Git Repository (Gitlab/GitHub)")
	downloadCmd.MarkFlagRequired("pat")
	viper.BindPFlag("pat", downloadCmd.Flags().Lookup("pat"))

	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Downloads all CREATE repositories via Git",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		Download(args)
	},
}
