package utils

import (
	"encoding/base64"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
	"sort"
	"strings"
	"testing"
)

func TestCopyMetricFamily(t *testing.T) {
	name := "test0"
	f := &dto.MetricFamily{
		Name: &name,
	}

	newName := "test1"
	f2 := CopyMetricFamily(f)
	f2.Name = &newName

	// test if f2 is really a deep copy or
	// if we still have references to values of f
	if *f.Name != name {
		t.Errorf("copied pointer not value, got: %s, expected: %s", *f.Name, name)
	}
}

func TestDecodeBase64(t *testing.T) {
	testStr := "test"
	testEncoded := base64.RawURLEncoding.EncodeToString([]byte(testStr))

	str, err := DecodeBase64(testEncoded)

	if err != nil {
		t.Error(err)
	}
	if str != testStr {
		t.Errorf("could not decode correctly, got: %s, expected: %s", str, testStr)
	}
}

func TestSplitLabels(t *testing.T) {
	labels := "key0/val0/key1/val1"

	comp, err := SplitLabels(labels, "")
	if err != nil {
		t.Error(err)
	}
	if len(comp) != 2 {
		t.Errorf("split components too small, got: %d, expected: %d", len(comp), 2)
	}

	labels1 := "key0/val0/key1"

	comp, err = SplitLabels(labels1, "")
	if err == nil {
		t.Errorf("expected odd number of components to fail, but it did not.")
	}

	labels2 := model.ReservedLabelPrefix + "key0/val0"

	comp, err = SplitLabels(labels2, "")
	if err == nil {
		t.Errorf("expected invalid label name to fail, but it did not.")
	}

	val := "val0"
	valEncoded := base64.RawURLEncoding.EncodeToString([]byte(val))
	labels3 := "key0@base64/" + valEncoded

	comp, err = SplitLabels(labels3, "@base64")
	if err != nil {
		t.Error(err)
	}
}

func TestGroupingKeyFor(t *testing.T) {
	str := GroupingKeyFor(nil)
	if str != "" {
		t.Errorf("expected nil value to return empty string, got: %s", str)
	}

	labels := make(map[string]string)
	str = GroupingKeyFor(labels)
	if str != "" {
		t.Errorf("expected empty map to return empty string, got: %s", str)
	}

	labels["key0"] = "value0"
	labels["key1"] = "value1"
	str = GroupingKeyFor(labels)
	if str == "" {
		t.Errorf("expected a valid grouping key, but got empty string")
	}
	split := strings.Split(str, string([]byte{model.SeparatorByte}))
	if len(split) != 4 {
		t.Errorf("expected length of splitted grouping key to be %d, but got %d", 4, len(split))
	}
}

func TestGroupingKeyForLabelPair(t *testing.T) {
	n := "key0"
	v := "value0"
	labelPairs := []*dto.LabelPair{
		{Name: &n, Value: &v},
	}

	str := GroupingKeyForLabelPair(labelPairs)
	if str == "" {
		t.Errorf("expected a valid grouping key, but got empty string")
	}
	split := strings.Split(str, string([]byte{model.SeparatorByte}))
	if len(split) != 2 {
		t.Errorf("expected length of splitted grouping key to be %d, but got %d", 2, len(split))
	}
}

func TestSanitizeLabels(t *testing.T) {
	groupingLabels := make(map[string]string)
	groupingLabels["gk0"] = "value0"
	groupingLabels["gk1"] = "value1"

	n := "name"
	mf := &dto.MetricFamily{
		Name: &n,
		Metric: []*dto.Metric{
			&dto.Metric{},
		},
	}

	SanitizeLabels(mf, groupingLabels)

	labels := mf.Metric[0].Label
	if len(labels) != 3 {
		t.Errorf("expected sanitize labels to append grouping labels, expected: %d, got: %d", 3, len(labels))
	}
	if *labels[0].Name != "gk0" || *labels[1].Name != "gk1" || *labels[2].Name != "instance" {
		t.Errorf("expected sanitize labels to append missing grouping labels, found: %v", labels)
	}
}

func TestTimestampsPresent(t *testing.T) {
	ts := int64(14)
	m := make(map[string]*dto.MetricFamily)

	mf := &dto.MetricFamily{
		Metric: []*dto.Metric{
			&dto.Metric{},
		},
	}
	m["0"] = mf

	if TimestampsPresent(m) {
		t.Errorf("expected no timestamps in map")
	}

	mf1 := &dto.MetricFamily{
		Metric: []*dto.Metric{
			&dto.Metric{
				TimestampMs: &ts,
			},
		},
	}
	m["1"] = mf1
	if !TimestampsPresent(m) {
		t.Errorf("expected a timestamp in map")
	}
}

func TestLabelPairs_Less(t *testing.T) {
	n, n1, n2 := "b", "a", "z"
	v, v1, v2 := "value0", "value1", "value2"
	labelPairs := LabelPairs([]*dto.LabelPair{
		{Name: &n, Value: &v},
		{Name: &n1, Value: &v1},
		{Name: &n2, Value: &v2},
	})

	sort.Sort(labelPairs)

	if *labelPairs[0].Name != "a" {
		t.Errorf("expected 'a' to be first after sorting, got: %s", *labelPairs[0].Name)
	}
}
