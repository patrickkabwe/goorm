package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/patrickkabwe/goorm"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Aliases: []string{"-m"},
	Short:   "Generate a migration file",
	Long:    "Generate a migration file",
	Example: "goorm migrate --name create_tables",
	Run: func(cmd *cobra.Command, args []string) {
		name := cmd.Flag("name").Value.String()
		if name == "" {
			fmt.Println("Please provide a name for the migration")
			os.Exit(1)
			return
		}
		migrate(name)
	},
}

func init() {
	migrateCmd.Flags().StringP("name", "n", "", "name of the migration")
	rootCmd.AddCommand(migrateCmd)
}

func migrate(migrationName string) {
	db, err := sql.Open(
		string(goorm.Postgres),
		testDB,
	)
	if err != nil {
		panic(err)
	}
	err = NewSchemaPusher(db, goorm.Postgres).Migrate(migrationName)
	if err != nil {
		panic(err)
	}
}
