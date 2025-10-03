package core

import (
	"fmt"

	"github.com/spf13/cobra"
)

func GetSendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Forward logs to the server",
		Long:  "Forward logs to the server.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Sending logs to the server...")
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	return cmd
}
