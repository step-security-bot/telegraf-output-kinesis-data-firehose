package firehose

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsfirehose "github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/muhlba91/telegraf-output-kinesis-data-firehose/serializer"
	"github.com/muhlba91/telegraf-output-kinesis-data-firehose/serializer/json"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

const (
	testRecordID         = "1"
	testStreamName       = "streamName"
	testBatchSize        = 2
	zero           int64 = 0
)

func TestWriteFirehose_WhenSuccess(t *testing.T) {
	records := []types.Record{
		{
			Data: []byte{0x65},
		},
	}

	svc := &mockFirehosePutRecordBatch{}
	svc.SetupResponse(
		0,
		[]types.PutRecordBatchResponseEntry{
			{
				RecordId: aws.String(testRecordID),
			},
		},
	)

	k := Output{
		StreamName: testStreamName,
		BatchSize:  maxRecordsPerRequest,
		svc:        svc,
	}

	elapsed := k.writeFirehose(records)
	require.GreaterOrEqual(t, elapsed.Nanoseconds(), zero)

	svc.AssertRequests(t, []*awsfirehose.PutRecordBatchInput{
		{
			DeliveryStreamName: aws.String(testStreamName),
			Records:            records,
		},
	})
}

func TestWriteFirehose_WhenRecordErrors(t *testing.T) {
	records := []types.Record{
		{
			Data: []byte{0x66},
		},
	}

	svc := &mockFirehosePutRecordBatch{}
	svc.SetupResponse(
		1,
		[]types.PutRecordBatchResponseEntry{
			{
				ErrorCode:    aws.String("InternalFailure"),
				ErrorMessage: aws.String("Internal Service Failure"),
			},
		},
	)

	k := Output{
		StreamName: testStreamName,
		BatchSize:  maxRecordsPerRequest,
		svc:        svc,
	}

	elapsed := k.writeFirehose(records)
	require.GreaterOrEqual(t, elapsed.Nanoseconds(), zero)

	svc.AssertRequests(t, []*awsfirehose.PutRecordBatchInput{
		{
			DeliveryStreamName: aws.String(testStreamName),
			Records:            records,
		},
	})
}

func TestWriteFirehose_WhenServiceError(t *testing.T) {
	records := []types.Record{
		{
			Data: []byte{},
		},
	}

	svc := &mockFirehosePutRecordBatch{}
	svc.SetupErrorResponse(
		&types.InvalidArgumentException{Message: aws.String("Invalid record")},
	)

	k := Output{
		StreamName: testStreamName,
		BatchSize:  maxRecordsPerRequest,
		svc:        svc,
	}

	elapsed := k.writeFirehose(records)
	require.GreaterOrEqual(t, elapsed.Nanoseconds(), zero)

	svc.AssertRequests(t, []*awsfirehose.PutRecordBatchInput{
		{
			DeliveryStreamName: aws.String(testStreamName),
			Records:            records,
		},
	})
}

func TestWrite_NoMetrics(t *testing.T) {
	serializer := createSerializer()
	svc := &mockFirehosePutRecordBatch{}

	k := Output{
		StreamName: "stream",
		BatchSize:  maxRecordsPerRequest,
		serializer: serializer,
		svc:        svc,
	}

	err := k.Write([]telegraf.Metric{})
	require.NoError(t, err, "Should not return error")

	svc.AssertRequests(t, []*awsfirehose.PutRecordBatchInput{})
}

func TestWrite_SingleMetric(t *testing.T) {
	serializer := createSerializer()

	svc := &mockFirehosePutRecordBatch{}
	svc.SetupGenericResponse(1, 0)

	k := Output{
		StreamName: testStreamName,
		BatchSize:  maxRecordsPerRequest,
		serializer: serializer,
		svc:        svc,
	}

	metric, metricData := createTestMetric(t, "metric1", serializer)
	err := k.Write([]telegraf.Metric{metric})
	require.NoError(t, err, "Should not return error")

	svc.AssertRequests(t, []*awsfirehose.PutRecordBatchInput{
		{
			DeliveryStreamName: aws.String(testStreamName),
			Records: []types.Record{
				{
					Data: metricData,
				},
			},
		},
	})
}

func TestWrite_MultipleMetrics_SinglePartialRequest(t *testing.T) {
	serializer := createSerializer()

	svc := &mockFirehosePutRecordBatch{}
	svc.SetupGenericResponse(3, 0)

	k := Output{
		StreamName: testStreamName,
		BatchSize:  maxRecordsPerRequest,
		serializer: serializer,
		svc:        svc,
	}

	metrics, metricsData := createTestMetrics(t, 3, serializer)
	err := k.Write(metrics)
	require.NoError(t, err, "Should not return error")

	svc.AssertRequests(t, []*awsfirehose.PutRecordBatchInput{
		{
			DeliveryStreamName: aws.String(testStreamName),
			Records: createRecordEntries(
				metricsData,
			),
		},
	})
}

func TestWrite_MultipleMetrics_SingleFullRequest(t *testing.T) {
	serializer := createSerializer()

	svc := &mockFirehosePutRecordBatch{}
	svc.SetupGenericResponse(maxRecordsPerRequest, 0)

	k := Output{
		StreamName: testStreamName,
		BatchSize:  maxRecordsPerRequest,
		serializer: serializer,
		svc:        svc,
	}

	metrics, metricsData := createTestMetrics(t, maxRecordsPerRequest, serializer)
	err := k.Write(metrics)
	require.NoError(t, err, "Should not return error")

	svc.AssertRequests(t, []*awsfirehose.PutRecordBatchInput{
		{
			DeliveryStreamName: aws.String(testStreamName),
			Records: createRecordEntries(
				metricsData,
			),
		},
	})
}

func TestWrite_MultipleMetrics_MultipleRequests(t *testing.T) {
	serializer := createSerializer()

	svc := &mockFirehosePutRecordBatch{}
	svc.SetupGenericResponse(maxRecordsPerRequest, 0)
	svc.SetupGenericResponse(1, 0)

	k := Output{
		StreamName: testStreamName,
		BatchSize:  maxRecordsPerRequest,
		serializer: serializer,
		svc:        svc,
	}

	metrics, metricsData := createTestMetrics(t, maxRecordsPerRequest+1, serializer)
	err := k.Write(metrics)
	require.NoError(t, err, "Should not return error")

	svc.AssertRequests(t, []*awsfirehose.PutRecordBatchInput{
		{
			DeliveryStreamName: aws.String(testStreamName),
			Records: createRecordEntries(
				metricsData[0:maxRecordsPerRequest],
			),
		},
		{
			DeliveryStreamName: aws.String(testStreamName),
			Records: createRecordEntries(
				metricsData[maxRecordsPerRequest:],
			),
		},
	})
}

func TestWrite_MultipleMetrics_MultipleFullRequests(t *testing.T) {
	serializer := createSerializer()

	svc := &mockFirehosePutRecordBatch{}
	svc.SetupGenericResponse(maxRecordsPerRequest, 0)
	svc.SetupGenericResponse(maxRecordsPerRequest, 0)

	k := Output{
		StreamName: testStreamName,
		BatchSize:  maxRecordsPerRequest,
		serializer: serializer,
		svc:        svc,
	}

	metrics, metricsData := createTestMetrics(t, maxRecordsPerRequest*2, serializer)
	err := k.Write(metrics)
	require.NoError(t, err, "Should not return error")

	svc.AssertRequests(t, []*awsfirehose.PutRecordBatchInput{
		{
			DeliveryStreamName: aws.String(testStreamName),
			Records: createRecordEntries(
				metricsData[0:maxRecordsPerRequest],
			),
		},
		{
			DeliveryStreamName: aws.String(testStreamName),
			Records: createRecordEntries(
				metricsData[maxRecordsPerRequest:],
			),
		},
	})
}

func TestWrite_MultipleMetrics_MultipleRequests_BatchSize(t *testing.T) {
	serializer := createSerializer()

	svc := &mockFirehosePutRecordBatch{}
	svc.SetupGenericResponse(testBatchSize, 0)
	svc.SetupGenericResponse(1, 0)

	k := Output{
		StreamName: testStreamName,
		BatchSize:  testBatchSize,
		serializer: serializer,
		svc:        svc,
	}

	metrics, metricsData := createTestMetrics(t, testBatchSize+1, serializer)
	err := k.Write(metrics)
	require.NoError(t, err, "Should not return error")

	svc.AssertRequests(t, []*awsfirehose.PutRecordBatchInput{
		{
			DeliveryStreamName: aws.String(testStreamName),
			Records: createRecordEntries(
				metricsData[0:testBatchSize],
			),
		},
		{
			DeliveryStreamName: aws.String(testStreamName),
			Records: createRecordEntries(
				metricsData[testBatchSize:],
			),
		},
	})
}

func TestWrite_SerializerError(t *testing.T) {
	serializer := createSerializer()

	svc := &mockFirehosePutRecordBatch{}
	svc.SetupGenericResponse(2, 0)

	k := Output{
		StreamName: testStreamName,
		BatchSize:  maxRecordsPerRequest,
		serializer: serializer,
		svc:        svc,
	}

	metric1, metric1Data := createTestMetric(t, "metric1", serializer)
	metric2, metric2Data := createTestMetric(t, "metric2", serializer)

	// metric is invalid because of empty name
	invalidMetric := testutil.TestMetric(3, "")

	err := k.Write([]telegraf.Metric{
		metric1,
		invalidMetric,
		metric2,
	})
	require.NoError(t, err, "Should not return error")

	// remaining valid metrics should still get written
	svc.AssertRequests(t, []*awsfirehose.PutRecordBatchInput{
		{
			DeliveryStreamName: aws.String(testStreamName),
			Records: []types.Record{
				{
					Data: metric1Data,
				},
				{
					Data: metric2Data,
				},
			},
		},
	})
}

func TestLoadFromConfig(t *testing.T) {
	k, err := NewOutput("test_files/minimal.conf")
	require.NoError(t, err, "Should not return error")
	require.Equal(t, k.StreamName, "streamname")
	require.Equal(t, k.Debug, false)
	require.Equal(t, k.BatchSize, maxRecordsPerRequest)

	k, err = NewOutput("test_files/debug.conf")
	require.NoError(t, err, "Should not return error")
	require.Equal(t, k.StreamName, "streamname")
	require.Equal(t, k.Debug, true)
	require.Equal(t, k.BatchSize, maxRecordsPerRequest)

	_, err = NewOutput("test_files/empty.conf")
	require.Error(t, err, "Should return error")

	_, err = NewOutput("test_files/not_existing.conf")
	require.Error(t, err, "Should return error")
}

func TestLoadFromConfig_BatchSize(t *testing.T) {
	k, err := NewOutput("test_files/batchsize.conf")
	require.NoError(t, err, "Should not return error")
	require.Equal(t, k.BatchSize, uint32(5))

	k, err = NewOutput("test_files/batchsize_too_large.conf")
	require.NoError(t, err, "Should not return error")
	require.Equal(t, k.BatchSize, maxRecordsPerRequest)

	_, err = NewOutput("test_files/batchsize_negative.conf")
	require.Error(t, err, "Should return error")
}

func TestLoadFromConfig_Format(t *testing.T) {
	k, err := NewOutput("test_files/minimal.conf")
	require.NoError(t, err, "Should not return error")
	require.Equal(t, k.Format.TimestampAsRFC3339, false)

	k, err = NewOutput("test_files/format_timestamp_as_rfc3339.conf")
	require.NoError(t, err, "Should not return error")
	require.Equal(t, k.Format.TimestampAsRFC3339, true)

	k, err = NewOutput("test_files/format_timestamp_units.conf")
	require.NoError(t, err, "Should not return error")
	require.Equal(t, k.Format.TimestampUnits, "1ms")

	_, err = NewOutput("test_files/format_invalid_timestamp_units.conf")
	require.Error(t, err, "Should return error")

	_, err = NewOutput("test_files/format_invalid.conf")
	require.Error(t, err, "Should return error")
}

func TestClose(t *testing.T) {
	k, _ := NewOutput("test_files/minimal.conf")
	err := k.Close()
	require.NoError(t, err, "Should not return error")
}

type mockFirehosePutRecordBatchResponse struct {
	Output *awsfirehose.PutRecordBatchOutput
	Err    error
}

type mockFirehosePutRecordBatch struct {
	requests  []*awsfirehose.PutRecordBatchInput
	responses []*mockFirehosePutRecordBatchResponse
}

func (m *mockFirehosePutRecordBatch) SetupResponse(
	failedRecordCount int32,
	records []types.PutRecordBatchResponseEntry,
) {
	m.responses = append(m.responses, &mockFirehosePutRecordBatchResponse{
		Err: nil,
		Output: &awsfirehose.PutRecordBatchOutput{
			FailedPutCount:   aws.Int32(failedRecordCount),
			RequestResponses: records,
		},
	})
}

func (m *mockFirehosePutRecordBatch) SetupGenericResponse(
	successfulRecordCount uint32,
	failedRecordCount int32,
) {
	records := []types.PutRecordBatchResponseEntry{}

	for i := uint32(0); i < successfulRecordCount; i++ {
		records = append(records, types.PutRecordBatchResponseEntry{
			RecordId: aws.String(testRecordID),
		})
	}

	for i := int32(0); i < failedRecordCount; i++ {
		records = append(records, types.PutRecordBatchResponseEntry{
			ErrorCode:    aws.String("InternalFailure"),
			ErrorMessage: aws.String("Internal Service Failure"),
		})
	}

	m.SetupResponse(failedRecordCount, records)
}

func (m *mockFirehosePutRecordBatch) SetupErrorResponse(err error) {
	m.responses = append(m.responses, &mockFirehosePutRecordBatchResponse{
		Err:    err,
		Output: nil,
	})
}

func (m *mockFirehosePutRecordBatch) PutRecordBatch(
	_ context.Context,
	input *awsfirehose.PutRecordBatchInput,
	_ ...func(*awsfirehose.Options),
) (*awsfirehose.PutRecordBatchOutput, error) {
	reqNum := len(m.requests)
	if reqNum > len(m.responses) {
		return nil, fmt.Errorf("response for request %+v not setup", reqNum)
	}

	m.requests = append(m.requests, input)

	resp := m.responses[reqNum]
	return resp.Output, resp.Err
}

func (m *mockFirehosePutRecordBatch) AssertRequests(
	t *testing.T,
	expected []*awsfirehose.PutRecordBatchInput,
) {
	require.Equalf(t,
		len(expected),
		len(m.requests),
		"Expected %v requests", len(expected),
	)

	for i, expectedInput := range expected {
		actualInput := m.requests[i]

		require.Equalf(t,
			expectedInput.DeliveryStreamName,
			actualInput.DeliveryStreamName,
			"Expected request %v to have correct DeliveryStreamName", i,
		)

		require.Equalf(t,
			len(expectedInput.Records),
			len(actualInput.Records),
			"Expected request %v to have %v Records", i, len(expectedInput.Records),
		)

		for r, expectedRecord := range expectedInput.Records {
			actualRecord := actualInput.Records[r]
			require.Equalf(t,
				expectedRecord.Data,
				actualRecord.Data,
				"Expected (request %v, record %v) to have correct Data", i, r,
			)
		}
	}
}

func createTestMetric(
	t *testing.T,
	name string,
	serializer *json.Serializer,
) (telegraf.Metric, []byte) {
	metric := testutil.TestMetric(1, name)

	data, err := serializer.Serialize(metric)
	require.NoError(t, err)

	return metric, data
}

func createTestMetrics(
	t *testing.T,
	count uint32,
	serializer *json.Serializer,
) ([]telegraf.Metric, [][]byte) {
	metrics := make([]telegraf.Metric, count)
	metricsData := make([][]byte, count)

	for i := uint32(0); i < count; i++ {
		name := fmt.Sprintf("metric%d", i)
		metric, data := createTestMetric(t, name, serializer)
		metrics[i] = metric
		metricsData[i] = data
	}

	return metrics, metricsData
}

func createRecordEntries(
	metricsData [][]byte,
) []types.Record {
	count := len(metricsData)
	records := make([]types.Record, count)

	for i := 0; i < count; i++ {
		records[i] = types.Record{
			Data: metricsData[i],
		}
	}

	return records
}

func createSerializer() *json.Serializer {
	serializer, _ := json.NewSerializer(time.Nanosecond, time.RFC3339, &serializer.Formatter{
		Flatten:       false,
		NameKeyRename: "",
		NormalizeKeys: false,
	})
	return serializer
}
