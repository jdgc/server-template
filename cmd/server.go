package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"jdgc/lists-server/v2/db"
	"jdgc/lists-server/v2/internal/lists"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "web server",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "start the web server",
	Run: func(cmd *cobra.Command, args []string) {
		db.Init()

		port, err := cmd.Flags().GetString("port")
		if err != nil {
			fmt.Println("Unable to read flag `name`", err.Error())
			return
		}

		listHandlers := lists.NewListHandlers()

		http.HandleFunc("/lists", listHandlers.Lists)
		http.HandleFunc("lists/", listHandlers.GetList)

		log.Printf("Server starting on port %s", port)
		serverPort := ":" + port
		err = http.ListenAndServe(serverPort, nil)

		if err != nil {
			panic(err)
		}
	},
}

func init() {
	serverStartCmd.Flags().StringP("port", "p", "8080", "port")

	serverCmd.AddCommand(serverStartCmd)

	rootCmd.AddCommand(serverCmd)
}
