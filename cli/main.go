package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "quest",
		Short: "Quest CLI",
		Long:  `CLI to manage firecracker microVMs`,
	}

	rootCmd.AddCommand(initCmd, startCmd, stopCmd, statusCmd, listCmd, deleteCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
