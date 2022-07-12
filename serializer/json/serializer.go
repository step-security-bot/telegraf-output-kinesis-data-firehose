package json

import (
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/muhlba91/telegraf-output-kinesis-data-firehose/serializer"

	"github.com/influxdata/telegraf"
)

type Serializer struct {
	TimestampUnits  time.Duration
	TimestampFormat string
	Formatter       *serializer.Formatter
}

func NewSerializer(timestampUnits time.Duration, timestampFormat string, formatter *serializer.Formatter) (*Serializer, error) {
	s := &Serializer{
		TimestampUnits:  truncateDuration(timestampUnits),
		TimestampFormat: timestampFormat,
		Formatter:       formatter,
	}
	return s, nil
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	m := s.createObject(metric)
	serialized, err := json.Marshal(m)
	if err != nil {
		return []byte{}, err
	}
	serialized = append(serialized, '\n')

	return serialized, nil
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	objects := make([]interface{}, 0, len(metrics))
	for _, metric := range metrics {
		m := s.createObject(metric)
		objects = append(objects, m)
	}

	obj := map[string]interface{}{
		"metrics": objects,
	}

	serialized, err := json.Marshal(obj)
	if err != nil {
		return []byte{}, err
	}
	return serialized, nil
}

func (s *Serializer) createObject(metric telegraf.Metric) map[string]interface{} {
	m := make(map[string]interface{}, 4)

	s.appendTags(&m, metric.TagList())
	s.appendFields(&m, metric.FieldList())

	s.appendName(&m, metric.Name())

	if s.TimestampFormat == "" {
		m["timestamp"] = metric.Time().UnixNano() / int64(s.TimestampUnits)
	} else {
		m["timestamp"] = metric.Time().UTC().Format(s.TimestampFormat)
	}
	return m
}

func (s *Serializer) appendTags(data *map[string]interface{}, tags []*telegraf.Tag) {
	d := *data
	if !s.Formatter.Flatten {
		d = make(map[string]interface{}, len(tags))
		(*data)["tags"] = d
	}
	for _, tag := range tags {
		d[s.normalizeKey(tag.Key)] = tag.Value
	}
}

func (s *Serializer) appendFields(data *map[string]interface{}, fields []*telegraf.Field) {
	d := *data
	if !s.Formatter.Flatten {
		d = make(map[string]interface{}, len(fields))
		(*data)["fields"] = d
	}
	for _, field := range fields {
		if fv, ok := field.Value.(float64); ok {
			// JSON does not support these special values
			if math.IsNaN(fv) || math.IsInf(fv, 0) {
				continue
			}
		}
		d[s.normalizeKey(field.Key)] = field.Value
	}
}

func (s *Serializer) appendName(data *map[string]interface{}, name string) {
	n := "name"
	if len(s.Formatter.NameKeyRename) > 0 {
		n = s.Formatter.NameKeyRename
	}
	(*data)[n] = name
}

func (s *Serializer) normalizeKey(key string) string {
	k := key
	if s.Formatter.NormalizeKeys {
		k = strings.ReplaceAll(strings.ToLower(key), " ", "_")
	}
	return k
}

func truncateDuration(units time.Duration) time.Duration {
	// Default precision is 1s
	if units <= 0 {
		return time.Second
	}

	// Search for the power of ten less than the duration
	d := time.Nanosecond
	for {
		if d*10 > units {
			return d
		}
		d = d * 10
	}
}
