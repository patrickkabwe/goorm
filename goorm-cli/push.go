package main

import (
	"database/sql"
	"fmt"

	"github.com/patrickkabwe/goorm"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:     "push",
	Aliases: []string{"p"},
	Short:   "Pushes the models to the database",
	Long:    "Pushes the models to the database that generated with the generate command",
	Run: func(cmd *cobra.Command, args []string) {
		push()
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func push() {
	db, err := sql.Open(
		string(goorm.Postgres),
		testDB,
	)

	if err != nil {
		panic(err)
	}
	err = NewSchemaPusher(db, goorm.Postgres).Push()

	if err != nil {
		panic(fmt.Sprintf("❌ Failed to push your models to the database: %v", err))
	}

	fmt.Println("✅ Your models have been pushed to the database")
}
