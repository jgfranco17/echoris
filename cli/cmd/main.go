package main

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jgfranco17/echoris/cli/core"

	_ "embed" // Required for the //go:embed directive
)

//go:embed specs.json
var embeddedConfig []byte

type ProjectMetadata struct {
	Author      string `json:"author"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Repository  string `json:"repository"`
}

func main() {
	var metadata ProjectMetadata
	if err := json.Unmarshal(embeddedConfig, &metadata); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
		os.Exit(1)
	}

	command := core.NewCommandRegistry(metadata.Name, metadata.Description, metadata.Version)
	commandsList := []*cobra.Command{
		core.GetSendCommand(),
	}
	command.RegisterCommands(commandsList)

	if err := command.Execute(); err != nil {
		log.Error(err.Error())
	}
}
