package firehose

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsfirehose "github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/plugins/common/aws"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/muhlba91/telegraf-output-kinesis-data-firehose/serializer"
	"github.com/muhlba91/telegraf-output-kinesis-data-firehose/serializer/json"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//
//go:embed sample.conf
var sampleConfig string

// Limit set by AWS (https://docs.aws.amazon.com/firehose/latest/APIReference/API_PutRecordBatch.html).
const maxRecordsPerRequest uint32 = 500

type (
	Firehose struct {
		StreamName string               `toml:"streamname"`
		Debug      bool                 `toml:"debug"`
		BatchSize  uint32               `toml:"batchsize"`
		Format     serializer.Formatter `toml:"format"`

		serializer *json.Serializer
		svc        Client

		internalaws.CredentialConfig
	}
)

type Client interface {
	PutRecordBatch(context.Context,
		*awsfirehose.PutRecordBatchInput,
		...func(*awsfirehose.Options)) (*awsfirehose.PutRecordBatchOutput, error)
	DescribeDeliveryStream(ctx context.Context,
		params *awsfirehose.DescribeDeliveryStreamInput,
		optFns ...func(*awsfirehose.Options)) (*awsfirehose.DescribeDeliveryStreamOutput, error)
}

func (f *Firehose) SampleConfig() string {
	return sampleConfig
}

func (f *Firehose) Init() error {
	var err error

	// Validate time formatting options.
	tf := ""
	if f.Format.TimestampAsRFC3339 {
		tf = time.RFC3339
		f.debug("setting timestamp parsing to RFC3339")
	}
	tu := time.Millisecond
	if len(f.Format.TimestampUnits) > 0 {
		if tu, err = time.ParseDuration(f.Format.TimestampUnits); err != nil {
			return err
		}
		f.debug(fmt.Sprintf("setting timestamp units to %s", tu.String()))
	}

	// Initialize JSON serializer.
	serializer, err := json.NewSerializer(tu, tf, &f.Format)
	if err != nil {
		return err
	}
	f.serializer = serializer

	// Validate the batch size.
	if f.BatchSize <= 0 || f.BatchSize > maxRecordsPerRequest {
		return fmt.Errorf("provided batch size if not within range: 0 < X <= %d", maxRecordsPerRequest)
	}

	// Validate stream name.
	if len(f.StreamName) == 0 {
		return errors.New("no stream name set")
	}

	return nil
}

func (f *Firehose) Connect() error {
	f.debug(fmt.Sprintf("connecting to the Amazon Kinesis Data Firehose %s in region %s", f.StreamName, f.Region))

	// Fetch and verify AWS credentials.
	cfg, err := f.CredentialConfig.Credentials()
	if err != nil {
		return err
	}

	// Create firehose service, and verify the stream.
	svc := awsfirehose.NewFromConfig(cfg)
	_, err = svc.DescribeDeliveryStream(context.Background(), &awsfirehose.DescribeDeliveryStreamInput{
		DeliveryStreamName: aws.String(f.StreamName),
	})
	f.log(fmt.Sprintf(
		"verification of the connection to Amazon Kinesis Data Firehose %s in region %s resulted in (nil = success): %s",
		f.StreamName,
		f.Region,
		err,
	))
	f.svc = svc

	return err
}

func (f *Firehose) Close() error {
	return nil
}

func (f *Firehose) Write(metrics []telegraf.Metric) error {
	var sz uint32

	// Return if no metrics have been received.
	if len(metrics) == 0 {
		return nil
	}
	f.debug(fmt.Sprintf("received %d metrics to process", len(metrics)))

	// Collect records to send to firehose.
	r := []types.Record{}

	for _, metric := range metrics {
		sz++

		// Skip metrics that don't have a 'name' property set.
		if len(metric.Name()) == 0 {
			continue
		}

		// Serialize the metric into JSON.
		// We skip unserializable metrics as they are invalid data points.
		values, err := f.serializer.Serialize(metric)
		if err != nil {
			f.log(fmt.Sprintf("unable to serialize metric %s: %s", metric, err))
			continue
		}

		d := types.Record{
			Data: values,
		}
		r = append(r, d)

		// Handle batch upload.
		if sz >= f.BatchSize {
			if err = f.writeToFirehose(r); err != nil {
				return err
			}
			sz = 0
			r = []types.Record{}
		}
	}
	// Send all left-over records to firehose.
	if sz > 0 {
		if err := f.writeToFirehose(r); err != nil {
			return err
		}
	}

	return nil
}

func (f *Firehose) writeToFirehose(r []types.Record) error {
	start := time.Now()

	// Create the firehose payload
	payload := &awsfirehose.PutRecordBatchInput{
		Records:            r,
		DeliveryStreamName: aws.String(f.StreamName),
	}

	// Send data to firehose.
	resp, err := f.svc.PutRecordBatch(context.Background(), payload)
	if err != nil {
		f.log(fmt.Sprintf("unable to write data to Amazon Kinesis Data Firehose %s: %s", f.StreamName, err.Error()))
		return err
	}
	f.debug(fmt.Sprintf("wrote data to firehose and got response: %+v", resp))

	// Verify if all data has been acknowledged by firehose.
	failed := *resp.FailedPutCount
	if failed > 0 {
		f.log(fmt.Sprintf("unable to write %+v of %+v record(s) to Amazon Kinesis Data Firehose %s", failed, len(r), f.StreamName))
	}

	f.debug(fmt.Sprintf("wrote a batch of %d to Amazon Kinesis Data Firehose %s in %+v", len(r), f.StreamName, time.Since(start)))

	return nil
}

func (f *Firehose) debug(t string) {
	if f.Debug {
		f.log(t)
	}
}

func (f *Firehose) log(t string) {
	log.Print(t)
}

func init() {
	outputs.Add("firehose", func() telegraf.Output { return &Firehose{} })
}
