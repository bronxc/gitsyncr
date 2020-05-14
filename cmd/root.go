package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
)

const (
	rootCmd = &cobra.Command{
		Use:   "gitsyncr",
		Short: "Tool for syncing your forks",
		Long: `Simple tool that pulls latest changes from your fork's
upstream remote and pushes them to your fork's remote.`,
		Run: func(cmd *cobra.Command, args []string) {
			Gitsyncr()
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
