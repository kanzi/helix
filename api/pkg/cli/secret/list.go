package secret

import (
	"fmt"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/helixml/helix/api/pkg/client"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List helix secrets",
	Long:    ``,
	RunE: func(cmd *cobra.Command, _ []string) error {
		apiClient, err := client.NewClientFromEnv()
		if err != nil {
			return err
		}

		secrets, err := apiClient.ListSecrets(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list secrets: %w", err)
		}

		table := tablewriter.NewWriter(cmd.OutOrStdout())

		header := []string{"Name", "App ID", "Created", "Updated"}

		table.SetHeader(header)

		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetTablePadding(" ")
		table.SetNoWhiteSpace(false)

		for _, s := range secrets {
			row := []string{
				s.Name,
				s.AppID,
				s.Created.Format(time.RFC3339),
				s.Updated.Format(time.RFC3339),
			}

			table.Append(row)
		}

		table.Render()

		return nil
	},
}
