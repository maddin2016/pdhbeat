package perfmon

import (
	"errors"
	"strconv"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/metricbeat/mb"
)

type CounterConfig struct {
	Alias string `config:"alias"`
	Query string `config:"query"`
}

// init registers the MetricSet with the central registry.
// The New method will be called after the setup of the module and before starting to fetch data
func init() {
	if err := mb.Registry.AddMetricSet("windows", "perfmon", New); err != nil {
		panic(err)
	}
}

// MetricSet type defines all fields of the MetricSet
// As a minimum it must inherit the mb.BaseMetricSet fields, but can be extended with
// additional entries. These variables can be used to persist data or configuration between
// multiple fetch calls.
type MetricSet struct {
	mb.BaseMetricSet
	counters []CounterConfig
	handle   *Handle
}

// New create a new instance of the MetricSet
// Part of new is also setting up the configuration by processing additional
// configuration entries if needed.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {

	config := struct {
		CounterConfig []CounterConfig `config:"counters"`
	}{}

	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	for _, v := range config.CounterConfig {
		if len(v.Alias) <= 0 {
			err := errors.New("Alias for counter cannot be empty")
			logp.Err("%v", err)
			return nil, err
		}
	}

	query, err := GetHandle(config.CounterConfig)

	if err != 0 {
		logp.Err("%v", err)
		return nil, errors.New("Initialization fails with error: " + strconv.Itoa(err))
	}

	return &MetricSet{
		BaseMetricSet: base,
		counters:      config.CounterConfig,
		handle:        query,
	}, nil
}

// Fetch methods implements the data gathering and data conversion to the right format
// It returns the event which is then forward to the output. In case of an error, a
// descriptive error must be returned.
func (m *MetricSet) Fetch() (common.MapStr, error) {

	data, err := m.handle.ReadData()
	if err != 0 {
		logp.Err("%v", err)
		return nil, errors.New("Fetching fails wir error: " + strconv.Itoa(err))
	}

	event := common.MapStr{
		"data": data,
	}

	return event, nil
}
