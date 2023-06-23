// Note: file taken from https://github.com/influxdata/telegraf/blob/master/plugins/common/shim/example/cmd/main.go
package main

import (
	"flag"
	"log"
	"os"
	"time"

	_ "github.com/muhlba91/telegraf-output-kinesis-data-firehose/plugins/outputs/firehose"

	"github.com/influxdata/telegraf/plugins/common/shim"
)

var Version string

var pollInterval = flag.Duration("poll_interval", 1*time.Second, "how often to send metrics")

var pollIntervalDisabled = flag.Bool(
	"poll_interval_disabled",
	false,
	"set to true to disable polling",
)

var (
	configFile = flag.String("config", "", "path to the config file for this plugin")
	err        error
)

func main() {
	log.Printf("aws kinesis data firehose plugin version: %s", Version)

	// Parse command line options.
	flag.Parse()
	if *pollIntervalDisabled {
		*pollInterval = shim.PollIntervalDisabled
	}

	// Create the shim to run the plugin.
	shimLayer := shim.New()

	// If no config is specified, all imported plugins are loaded.
	// otherwise, follow what the config asks for.
	// Check for settings from a config toml file,
	// (or just use whatever plugins were imported above)
	if err = shimLayer.LoadConfig(configFile); err != nil {
		log.Printf("error loading plugin: %s", err)
		os.Exit(1)
	}

	// Run the plugin until stdin closes, or we receive a termination signal.
	if err = shimLayer.Run(*pollInterval); err != nil {
		log.Printf("error running plugin: %s", err)
		os.Exit(1)
	}
}
