package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"sm2/cli"
	"sm2/ledger"
	"sm2/platform"
	"sm2/servicemanager"
)

func main() {

	cmds, err := cli.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("Invalid option: %s\n", err)
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: 30 * time.Minute,
	}

	serviceManager := servicemanager.ServiceManager{
		Client:   client,
		Commands: *cmds,

		Platform: platform.DetectPlatform(),
		Ledger:   ledger.NewLedger(),
	}

	err = serviceManager.LoadConfig()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	serviceManager.Run()

	ts := platform.GetTerminalSize()
	fmt.Printf("Terminal Size cols %d rows %d\n", ts.Cols, ts.Rows)
}
