package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wallix/awless/cloud/aws"
	"github.com/wallix/awless/config"
	"github.com/wallix/awless/database"
	"github.com/wallix/awless/stats"
)

func initAwlessEnvFn(cmd *cobra.Command, args []string) {
	if err := config.InitAwlessEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "cannot init awless environment: %s\n", err)
		os.Exit(1)
	}
}

func initCloudServicesFn(cmd *cobra.Command, args []string) {
	initAwlessEnvFn(cmd, args)
	if err := aws.InitServices(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := database.InitDB(config.AwlessFirstInstall); err != nil {
		fmt.Fprintf(os.Stderr, "cannot init database: %s\n", err)
		os.Exit(1)
	}
	checkStatsFn()
}

func saveHistoryFn(cmd *cobra.Command, args []string) {
	db, close := database.Current()
	defer close()
	db.AddHistoryCommand(append(strings.Split(cmd.CommandPath(), " "), args...))
}

func checkStatsFn() {
	db, dbclose := database.Current()
	statsToSend := stats.CheckStatsToSend(db)
	dbclose()
	if statsToSend {
		go sendStats()
	}
}

func sendStats() {
	var err error
	localInfra, err := config.LoadInfraGraph()
	if err != nil {
		return
	}
	localAccess, err := config.LoadAccessGraph()
	if err != nil {
		return
	}

	db, dbclose := database.Current()
	if err := stats.SendStats(db, localInfra, localAccess); err != nil {
		db.AddLog(err.Error())
	}
	dbclose()
}