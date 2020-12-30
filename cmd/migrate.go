package cmd

import (
	"fmt"

	"jdgc/lists-server/v2/db"
	"jdgc/lists-server/v2/db/migrations"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migration",
	Short: "database migrations",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var migrateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new migration file",
	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Println("Unable to read flag `name`", err.Error())
			return
		}

		if err := migrations.Create(name); err != nil {
			fmt.Println("Unable to create migration", err.Error())
			return
		}
	},
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "apply pending migrations",
	Run: func(cmd *cobra.Command, args []string) {

		step, err := cmd.Flags().GetInt("step")
		if err != nil {
			fmt.Println("Invalid step value")
			return
		}

		db := db.DB

		migrator, err := migrations.Init(db)
		if err != nil {
			fmt.Println("Unable to fetch migrator")
			return
		}

		err = migrator.Up(step)
		if err != nil {
			fmt.Println("Unable to run migrations")
			return
		}
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "reverse migrations",
	Run: func(cmd *cobra.Command, args []string) {

		step, err := cmd.Flags().GetInt("step")
		if err != nil {
			fmt.Println("Invalid step value")
			return
		}

		db := db.DB

		migrator, err := migrations.Init(db)
		if err != nil {
			fmt.Println("Unable to fetch migrator")
			return
		}

		err = migrator.Down(step)
		if err != nil {
			fmt.Println("Unable to reverse migrations")
			return
		}
	},
}

func init() {
	migrateCreateCmd.Flags().StringP("name", "n", "", "migration name")

	migrateUpCmd.Flags().IntP("step", "s", 0, "Number of migrations to process")
	migrateDownCmd.Flags().IntP("step", "s", 0, "Number of migrations to process")

	migrateCmd.AddCommand(migrateCreateCmd, migrateUpCmd, migrateDownCmd)

	rootCmd.AddCommand(migrateCmd)
}
