package storage

import (
	"github.com/golang/protobuf/proto"
	dto "github.com/prometheus/client_model/go"
	"testing"
)

func metricTypePtr(val dto.MetricType) *dto.MetricType {
	return &val
}

func TestInsertingWithDifferentValues(t *testing.T) {
	ms := &MetricStorage{
		metricGroups: make(map[string]MetricGroup),
	}

	metrics := make(map[string]*dto.MetricFamily)

	gauge := &dto.Metric{
		Label: []*dto.LabelPair{},
		Gauge: &dto.Gauge{
			Value: proto.Float64(-13),
		},
	}

	metrics["f1Name"] = &dto.MetricFamily{
		Name: proto.String("f1Name"),
		Type: metricTypePtr(dto.MetricType_GAUGE),
		Metric: []*dto.Metric{
			gauge,
		},
	}

	labels := make(map[string]string)
	wr := WriteRequest{
		Labels: labels,
		MetricFamilies: metrics,
	}

	// ==========
	// test begin
	// ==========

	ms.processWriteRequest(wr)

	val := ms.GetMetricFamilies()[0].Metric[0].Gauge.Value
	if *val != -13 {
		t.Errorf("could not process metric, expected gauge value: %v, got: %v", -13, *val)
	}

	metrics2 := make(map[string]*dto.MetricFamily)
	gauge2 := &dto.Metric{
		Label: []*dto.LabelPair{},
		Gauge: &dto.Gauge{
			Value: proto.Float64(4),
		},
	}
	metrics2["f1Name"] = &dto.MetricFamily{
		Name: proto.String("f1Name"),
		Type: metricTypePtr(dto.MetricType_GAUGE),
		Metric: []*dto.Metric{
			gauge2,
		},
	}
	wr2 := WriteRequest{
		Labels: labels,
		MetricFamilies: metrics2,
	}

	ms.processWriteRequest(wr2)

	val = ms.GetMetricFamilies()[0].Metric[0].Gauge.Value
	if *val != 4 {
		t.Errorf("could not process metric, expected gauge value: %v, got: %v", 4, *val)
	}
}

func TestInsertingWithTimestamp(t *testing.T) {
	ms := &MetricStorage{
		metricGroups: make(map[string]MetricGroup),
	}
	metrics := make(map[string]*dto.MetricFamily)

	gauge := &dto.Metric{
		TimestampMs: proto.Int64(0),
		Label: []*dto.LabelPair{},
		Gauge: &dto.Gauge{
			Value: proto.Float64(-13),
		},
	}

	f1Name := "f1Name"
	f1Type := dto.MetricType_GAUGE
	metrics[f1Name] = &dto.MetricFamily{
		Name: &f1Name,
		Type: &f1Type,
		Metric: []*dto.Metric{
			gauge,
		},
	}

	labels := make(map[string]string)
	wr := WriteRequest{
		Labels: labels,
		MetricFamilies: metrics,
	}

	// ==========
	// test begin
	// ==========

	err := validateConsistency(ms, wr)
	if err == nil {
		t.Errorf("expected metric with timestamp to fail, but it did not.")
	}
}

func TestInsertingDuplicateSameTypeDifferentLabels(t *testing.T) {
	ms := &MetricStorage{
		metricGroups: make(map[string]MetricGroup),
	}

	gauge := &dto.Metric{
		Label: []*dto.LabelPair{
			&dto.LabelPair{
				Name: proto.String("key0"),
				Value: proto.String("val0"),
			},
		},
		Gauge: &dto.Gauge{
			Value: proto.Float64(0),
		},
	}
	metrics := make(map[string]*dto.MetricFamily)
	metrics["f1Name"] = &dto.MetricFamily{
		Name: proto.String("f1Name"),
		Type: metricTypePtr(dto.MetricType_GAUGE),
		Metric: []*dto.Metric{
			gauge,
		},
	}

	labels := make(map[string]string)
	wr := WriteRequest{
		Labels: labels,
		MetricFamilies: metrics,
	}

	// ==========
	// test begin
	// ==========

	ms.processWriteRequest(wr)

	gauge2 := &dto.Metric{
		Label: []*dto.LabelPair{
			&dto.LabelPair{
				Name: proto.String("key1"),
				Value: proto.String("val0"),
			},
		},
		Gauge: &dto.Gauge{
			Value: proto.Float64(0),
		},
	}
	metrics2 := make(map[string]*dto.MetricFamily)
	metrics2["f1Name"] = &dto.MetricFamily{
		Name: proto.String("f1Name"),
		Type: metricTypePtr(dto.MetricType_GAUGE),
		Metric: []*dto.Metric{
			gauge2,
		},
	}
	wr.MetricFamilies = metrics2

	ms.processWriteRequest(wr)

	if len(ms.metricGroups) != 1 {
		t.Errorf("storage created excessive groups, expected: 1, got: %d", len(ms.metricGroups))
	}
	size := len(ms.GetMetricFamilies())
	if size != 1 {
		t.Errorf("storage created excessive families, expected: 1, got: %d", size)
	}
}

func TestInsertingDuplicateDifferentType(t *testing.T) {
	ms := &MetricStorage{
		metricGroups: make(map[string]MetricGroup),
	}

	metrics := make(map[string]*dto.MetricFamily)

	gauge := &dto.Metric{
		Label: []*dto.LabelPair{},
		Gauge: &dto.Gauge{
			Value: proto.Float64(0),
		},
	}

	metrics["f1Name"] = &dto.MetricFamily{
		Name: proto.String("f1Name"),
		Type: metricTypePtr(dto.MetricType_GAUGE),
		Metric: []*dto.Metric{
			gauge,
		},
	}

	labels := make(map[string]string)
	wr := WriteRequest{
		Labels: labels,
		MetricFamilies: metrics,
	}

	// ==========
	// test begin
	// ==========

	ms.processWriteRequest(wr)

	counter := &dto.Metric{
		Label: []*dto.LabelPair{},
		Counter: &dto.Counter{
			Value: proto.Float64(0),
		},
	}

	metrics2 := make(map[string]*dto.MetricFamily)
	metrics2["f1Name"] = &dto.MetricFamily{
		Name: proto.String("f1Name"),
		Type: metricTypePtr(dto.MetricType_COUNTER),
		Metric: []*dto.Metric{
			counter,
		},
	}
	wr.MetricFamilies = metrics2

	err := validateConsistency(ms, wr)
	if err == nil {
		t.Errorf("expected metric with same name but different type to fail, but it did not.")
	}
}

func TestMergingCounter(t *testing.T) {
	ms := &MetricStorage{
		metricGroups: make(map[string]MetricGroup),
	}

	counter := &dto.Metric{
		Label: []*dto.LabelPair{},
		Counter: &dto.Counter{
			Value: proto.Float64(5),
		},
	}

	metrics := make(map[string]*dto.MetricFamily)
	metrics["f1Name"] = &dto.MetricFamily{
		Name: proto.String("f1Name"),
		Type: metricTypePtr(dto.MetricType_COUNTER),
		Metric: []*dto.Metric{
			counter,
		},
	}

	labels := make(map[string]string)
	wr := WriteRequest{
		Labels: labels,
		MetricFamilies: metrics,
	}

	// ==========
	// test begin
	// ==========

	ms.processWriteRequest(wr)

	counter2 := &dto.Metric{
		Label: []*dto.LabelPair{},
		Counter: &dto.Counter{
			Value: proto.Float64(4),
		},
	}
	metrics2 := make(map[string]*dto.MetricFamily)
	metrics2["f1Name"] = &dto.MetricFamily{
		Name: proto.String("f1Name"),
		Type: metricTypePtr(dto.MetricType_COUNTER),
		Metric: []*dto.Metric{
			counter2,
		},
	}
	wr2 := WriteRequest{
		Labels: labels,
		MetricFamilies: metrics2,
	}

	ms.processWriteRequest(wr2)

	val := ms.GetMetricFamilies()[0].Metric[0].Counter.Value
	if *val != 5 + 4 {
		t.Errorf("could not merge metric, expected counter value: %v, got: %v", 5 + 4, *val)
	}
}

func TestMergingHistogram(t *testing.T) {
	ms := &MetricStorage{
		metricGroups: make(map[string]MetricGroup),
	}

	hist := &dto.Metric{
		Label: []*dto.LabelPair{},
		Histogram: &dto.Histogram{
			SampleCount: proto.Uint64(2),
			SampleSum: proto.Float64(13),
			Bucket: []*dto.Bucket{
				&dto.Bucket{
					CumulativeCount: proto.Uint64(13),
					UpperBound: proto.Float64(20),
				},
			},
		},
	}

	metrics := make(map[string]*dto.MetricFamily)
	metrics["f1Name"] = &dto.MetricFamily{
		Name: proto.String("f1Name"),
		Type: metricTypePtr(dto.MetricType_HISTOGRAM),
		Metric: []*dto.Metric{
			hist,
		},
	}

	labels := make(map[string]string)
	wr := WriteRequest{
		Labels: labels,
		MetricFamilies: metrics,
	}

	// ==========
	// test begin
	// ==========

	ms.processWriteRequest(wr)

	hist2 := &dto.Metric{
		Label: []*dto.LabelPair{},
		Histogram: &dto.Histogram{
			SampleCount: proto.Uint64(3),
			SampleSum: proto.Float64(40),
			Bucket: []*dto.Bucket{
				&dto.Bucket{
					CumulativeCount: proto.Uint64(13),
					UpperBound: proto.Float64(20),
				},
				&dto.Bucket{
					CumulativeCount: proto.Uint64(27),
					UpperBound: proto.Float64(40),
				},
			},
		},
	}
	metrics2 := make(map[string]*dto.MetricFamily)
	metrics2["f1Name"] = &dto.MetricFamily{
		Name: proto.String("f1Name"),
		Type: metricTypePtr(dto.MetricType_HISTOGRAM),
		Metric: []*dto.Metric{
			hist2,
		},
	}
	wr2 := WriteRequest{
		Labels: labels,
		MetricFamilies: metrics2,
	}

	ms.processWriteRequest(wr2)

	resHist := ms.GetMetricFamilies()[0].Metric[0].Histogram
	if len(resHist.Bucket) != 2 {
		t.Errorf("expected %d buckets, got %d", 2, len(resHist.Bucket))
	}
	if *resHist.SampleSum != 53 {
		t.Errorf("could not merge metric, expected histogram sample sum: %v, got: %v", 53, *resHist.SampleSum)
	}
}

func TestDelete(t *testing.T) {
	ms := &MetricStorage{
		metricGroups: make(map[string]MetricGroup),
	}

	gauge := &dto.Metric{
		Label: []*dto.LabelPair{},
		Gauge: &dto.Gauge{
			Value: proto.Float64(-13),
		},
	}
	metrics := make(map[string]*dto.MetricFamily)
	metrics["f1Name"] = &dto.MetricFamily{
		Name: proto.String("f1Name"),
		Type: metricTypePtr(dto.MetricType_GAUGE),
		Metric: []*dto.Metric{
			gauge,
		},
	}

	labels := make(map[string]string)
	labels["job"] = "test0"
	wr := WriteRequest{
		Labels: labels,
		MetricFamilies: metrics,
	}

	// ==========
	// test begin
	// ==========

	ms.processWriteRequest(wr)

	wr2 := WriteRequest{
		Labels: labels,
	}

	ms.processWriteRequest(wr2)

	val := len(ms.GetMetricFamilies())
	if val != 0 {
		t.Errorf("metric could not be deleted, found: %d", val)
	}
}
