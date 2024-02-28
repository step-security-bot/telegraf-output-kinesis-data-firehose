package json

import (
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/muhlba91/telegraf-output-kinesis-data-firehose/serializer"
	"github.com/stretchr/testify/assert"
)

func TestNewSerializer(t *testing.T) {
	s, err := NewSerializer(time.Second, "2006-01-02T15:04:05Z07:00", &serializer.Formatter{})
	assert.NoError(t, err)
	assert.NotNil(t, s)
}

func TestSerialize(t *testing.T) {
	s, _ := NewSerializer(time.Second, "2006-01-02T15:04:05Z07:00", &serializer.Formatter{})
	metric := testutil.TestMetric("value", "test")
	j, err := s.Serialize(metric)
	assert.NoError(t, err)
	assert.Equal(t, `{"fields":{"value":"value"},"name":"test","tags":{"tag1":"value1"},"timestamp":"2009-11-10T23:00:00Z"}`, strings.TrimSpace(string(j))) //nolint:golint,lll
}

func TestSerializeBatch(t *testing.T) {
	s, _ := NewSerializer(time.Second, "2006-01-02T15:04:05Z07:00", &serializer.Formatter{})
	metric1 := testutil.TestMetric("valu1e", "test1")
	metric2 := testutil.TestMetric("value2", "test1")
	j, err := s.SerializeBatch([]telegraf.Metric{metric1, metric2})
	assert.NoError(t, err)
	assert.Equal(t, `{"metrics":[{"fields":{"value":"valu1e"},"name":"test1","tags":{"tag1":"value1"},"timestamp":"2009-11-10T23:00:00Z"},{"fields":{"value":"value2"},"name":"test1","tags":{"tag1":"value1"},"timestamp":"2009-11-10T23:00:00Z"}]}`, strings.TrimSpace(string(j))) //nolint:golint,lll
}

func TestCreateObject(t *testing.T) {
	s, _ := NewSerializer(time.Second, "2006-01-02T15:04:05Z07:00", &serializer.Formatter{})
	s.TimestampFormat = ""
	metric := testutil.TestMetric("value", "test")
	obj := s.createObject(metric)
	assert.Equal(t, "test", obj["name"])
	assert.Equal(t, map[string]interface{}{"tag1": "value1"}, obj["tags"])
	assert.Equal(t, map[string]interface{}{"value": "value"}, obj["fields"])
	assert.Equal(t, int64(1257894000), obj["timestamp"])

	// Test with custom timestamp format
	s.TimestampFormat = "2006-01-02 15:04:05"
	obj = s.createObject(metric)
	assert.Equal(t, "test", obj["name"])
	assert.Equal(t, map[string]interface{}{"tag1": "value1"}, obj["tags"])
	assert.Equal(t, map[string]interface{}{"value": "value"}, obj["fields"])
	assert.Equal(t, "2009-11-10 23:00:00", obj["timestamp"])
}

func TestAppendTags(t *testing.T) {
	s, _ := NewSerializer(time.Second, "2006-01-02T15:04:05Z07:00", &serializer.Formatter{})
	data := make(map[string]interface{})
	s.appendTags(&data, []*telegraf.Tag{{Key: "tag", Value: "value"}})
	assert.Equal(t, map[string]interface{}{"tags": map[string]interface{}{"tag": "value"}}, data)
}

func TestAppendFields(t *testing.T) {
	s, _ := NewSerializer(time.Second, "2006-01-02T15:04:05Z07:00", &serializer.Formatter{})
	data := make(map[string]interface{})
	s.appendFields(&data, []*telegraf.Field{{Key: "field", Value: 1.0}})
	assert.Equal(t, map[string]interface{}{"fields": map[string]interface{}{"field": 1.0}}, data)
}

func TestAppendName(t *testing.T) {
	s, _ := NewSerializer(time.Second, "2006-01-02T15:04:05Z07:00", &serializer.Formatter{})
	data := make(map[string]interface{})
	s.appendName(&data, "test")
	assert.Equal(t, map[string]interface{}{"name": "test"}, data)

	// Test with custom name key rename
	s.Formatter.NameKeyRename = "custom_name"
	data = make(map[string]interface{})
	s.appendName(&data, "test")
	assert.Equal(t, map[string]interface{}{"custom_name": "test"}, data)
}

func TestNormalizeKey(t *testing.T) {
	s, _ := NewSerializer(time.Second, "2006-01-02T15:04:05Z07:00", &serializer.Formatter{NormalizeKeys: true})
	assert.Equal(t, "test_key", s.normalizeKey("Test Key"))
}

func TestTruncateDuration(t *testing.T) {
	assert.Equal(t, 10*time.Second, truncateDuration(time.Minute))
	assert.Equal(t, time.Second, truncateDuration(time.Second))
	assert.Equal(t, time.Millisecond, truncateDuration(time.Millisecond))
	assert.Equal(t, time.Second, truncateDuration(-1))
}
