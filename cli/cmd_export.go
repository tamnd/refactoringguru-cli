package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func (a *App) exportCmd() *cobra.Command {
	var outFile string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export all design patterns as JSONL",
		Long: `export fetches the full design-patterns catalog and writes one JSON record
per line. Use --out to write to a file instead of stdout.

Examples:
  rg export > patterns.jsonl
  rg export --out patterns.jsonl
  rg export -o csv > patterns.csv`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			patterns, err := a.client.Patterns(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			if outFile != "" {
				f, err := os.Create(outFile)
				if err != nil {
					return codeError(exitError, err)
				}
				defer f.Close()
				r := a.newRendererTo(f)
				return r.Render(patterns)
			}
			return a.renderOrEmpty(patterns, len(patterns))
		},
	}
	cmd.Flags().StringVar(&outFile, "out", "", "write output to FILE instead of stdout")
	return cmd
}
