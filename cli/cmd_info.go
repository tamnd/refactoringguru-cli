package cli

import "github.com/spf13/cobra"

func (a *App) infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show refactoring.guru catalog statistics",
		Long: `info prints aggregate statistics: total patterns and the count per GoF
category (Creational, Structural, Behavioral).

Examples:
  rg info
  rg info -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			info, err := a.client.Stats(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render(info)
		},
	}
}
