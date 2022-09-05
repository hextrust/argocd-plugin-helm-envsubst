package cmd

import (
	"github.com/spf13/cobra"
	app "gitlab.int.hextech.io/technology/utils/cicd/argocd-helm-envsubst-plugin/internal"
)

var (
	renderPath  string
	logLocation string
)

func init() {
	renderCmd.PersistentFlags().StringVar(&renderPath, "path", "", "Path to the application")
	renderCmd.PersistentFlags().StringVar(&logLocation, "log-location", "", "Default to /tmp/argocd-helm-envsubst-plugin/")
	rootCmd.AddCommand(renderCmd)
}

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Similar to helm template .",
	Run: func(cmd *cobra.Command, args []string) {
		app.NewRenderer().RenderTemplate(renderPath, logLocation)
	},
}
