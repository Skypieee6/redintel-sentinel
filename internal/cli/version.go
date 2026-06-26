package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Skypieee6/redintel-sentinel/internal/version"
)

func newVersionCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print build version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := version.Get()
			if asJSON {
				b, err := json.MarshalIndent(info, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(b))
				return nil
			}
			fmt.Println(info.String())
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output version information as JSON")
	return cmd
}
