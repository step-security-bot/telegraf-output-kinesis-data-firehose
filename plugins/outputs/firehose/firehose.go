package firehose

import (
	"context"
	_ "embed"
	"io/ioutil"
	"log"
	"time"

	"github.com/BurntSushi/toml"
	awsfirehose "github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/muhlba91/telegraf-output-kinesis-data-firehose/serializer"
	"github.com/muhlba91/telegraf-output-kinesis-data-firehose/serializer/json"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

// Limit set by AWS (https://docs.aws.amazon.com/firehose/latest/APIReference/API_PutRecordBatch.html)
const maxRecordsPerRequest uint32 = 500

type (
	Output struct {
		StreamName string               `toml:"streamname"`
		Debug      bool                 `toml:"debug"`
		Format     serializer.Formatter `toml:"format"`

		serializer *json.Serializer
		svc        firehoseClient

		internalaws.CredentialConfig
	}
)

type firehoseClient interface {
	PutRecordBatch(context.Context, *awsfirehose.PutRecordBatchInput, ...func(*awsfirehose.Options)) (*awsfirehose.PutRecordBatchOutput, error)
}

func SampleConfig() string {
	return sampleConfig
}

func NewOutput(configFilePath string) (*Output, error) {
	cfgBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	cfgBytes = parseConfig(cfgBytes)

	output := &Output{}
	err = toml.Unmarshal(cfgBytes, output)
	if err != nil {
		return nil, err
	}

	tf := ""
	if output.Format.TimestampAsRFC3339 {
		tf = time.RFC3339
	}
	tu := time.Millisecond
	if len(output.Format.TimestampUnits) > 0 {
		if tu, err = time.ParseDuration(output.Format.TimestampUnits); err != nil {
			return nil, err
		}
	}
	serializer, err := json.NewSerializer(tu, tf, &output.Format)
	if err != nil {
		return nil, err
	}

	output.serializer = serializer

	return output, nil
}

func (k *Output) Connect() error {
	if k.Debug {
		log.Printf("Establishing a connection to Amazon Kinesis Data Firehose in %s", k.Region)
	}

	cfg, err := k.CredentialConfig.Credentials()
	if err != nil {
		return err
	}

	svc := awsfirehose.NewFromConfig(cfg)

	_, err = svc.DescribeDeliveryStream(context.Background(), &awsfirehose.DescribeDeliveryStreamInput{
		DeliveryStreamName: aws.String(k.StreamName),
	})
	k.svc = svc
	log.Printf("Connected to Amazon Kinesis Data Firehose %s in %s: %t", k.StreamName, k.Region, err == nil)

	return err
}

func (k *Output) Close() error {
	return nil
}

func (k *Output) Write(metrics []telegraf.Metric) error {
	var sz uint32

	if len(metrics) == 0 {
		return nil
	}
	log.Printf("Received %d metrics", len(metrics))

	r := []types.Record{}

	for _, metric := range metrics {
		sz++

		if len(metric.Name()) == 0 {
			continue
		}

		values, err := k.serializer.Serialize(metric)
		if err != nil {
			log.Printf("Unable to serialize metric %s: %s", metric, err)
			continue
		}

		d := types.Record{
			Data: values,
		}
		r = append(r, d)

		if sz == maxRecordsPerRequest {
			elapsed := k.writeFirehose(r)
			if k.Debug {
				log.Printf("Wrote a %d point batch to Amazon Kinesis Data Firehose %s in %+v.", sz, k.StreamName, elapsed)
			}
			sz = 0
			r = nil
		}
	}
	if sz > 0 {
		elapsed := k.writeFirehose(r)
		if k.Debug {
			log.Printf("Wrote a %d point batch to Amazon Kinesis Data Firehose %s in %+v.", sz, k.StreamName, elapsed)
		}
	}

	return nil
}

func (k *Output) writeFirehose(r []types.Record) time.Duration {
	start := time.Now()
	payload := &awsfirehose.PutRecordBatchInput{
		Records:            r,
		DeliveryStreamName: aws.String(k.StreamName),
	}

	resp, err := k.svc.PutRecordBatch(context.Background(), payload)
	if err != nil {
		log.Printf("Unable to write to Amazon Kinesis Data Firehose %s: %s", k.StreamName, err.Error())
		return time.Since(start)
	}

	if k.Debug {
		log.Printf("Wrote: '%+v'", resp)
	}

	failed := *resp.FailedPutCount
	if failed > 0 {
		log.Printf("Unable to write %+v of %+v record(s) to Amazon Kinesis Data Firehose %s.", failed, len(r), k.StreamName)
	}

	return time.Since(start)
}
