package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	influxparser "github.com/influxdata/telegraf/plugins/parsers/influx"

	"github.com/muhlba91/telegraf-output-kinesis-data-firehose/plugins/outputs/firehose"
)

var Version = "v0.0.0"

func main() {
	flagVersion := flag.Bool("v", false, "Shows the version of the plugin.")
	flagSampleConfig := flag.Bool("sample", false, "Prints a sample configuration to stdout.")
	flagConfigFile := flag.String("config", "kinesis_output.toml", "The location of the configuration file to use.")
	flag.Parse()

	if *flagVersion {
		fmt.Println(Version)
		return
	}

	if *flagSampleConfig {
		fmt.Println(firehose.SampleConfig())
		return
	}

	stdinBytes, err := io.ReadAll(bufio.NewReader(os.Stdin))
	if err != nil {
		terminate("Failed to read passed in data: %s", err)
	}

	if len(stdinBytes) == 0 {
		return
	}

	parser := influxparser.Parser{}
	initErr := parser.Init()
	if initErr != nil {
		terminate("Failed to instantiate the parser: %s", initErr)
	}

	metricList, err := parser.Parse(stdinBytes)
	if err != nil {
		terminate("Failed to convert the metrics: %s", err)
	}

	output, err := firehose.NewOutput(*flagConfigFile)
	if err != nil {
		terminate("Failed to make Amazon Kinesis Firehose adaptor. Error: %s", err)
	}
	err = output.Connect()
	if err != nil {
		terminate("Couldn't connect to Amazon Kinesis Firehose. Error: %s", err)
	}
	err = output.Write(metricList)
	if err != nil {
		terminate("Failed to write to Kinesis. Error: %s", err)
	}
}

func terminate(format string, args ...interface{}) {
	log.Printf(format, args...)
	os.Exit(10)
}
