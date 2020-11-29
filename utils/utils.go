package utils

import (
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
	"sort"
	"strings"
)

func CopyMetricFamily(mf *dto.MetricFamily) *dto.MetricFamily {
	return proto.Clone(mf).(*dto.MetricFamily)
	/*return &dto.MetricFamily{
		Name:   mf.Name,
		Help:   mf.Help,
		Type:   mf.Type,
		Metric: append([]*dto.Metric{}, mf.Metric...),
	}*/
}

// DecodeBase64 decodes the provided string using the “Base 64 Encoding with URL
// and Filename Safe Alphabet” (RFC 4648). Padding characters (i.e. trailing
// '=') are ignored.
//
// Source: github.com/prometheus/pushgateway
func DecodeBase64(s string) (string, error) {
	b, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
	return string(b), err
}

// SplitLabels splits a labels string into a label map mapping names to values.
// they should be in the following format:
//  /key0/value0/key1/value1/.../keyN/valueN
//
// Source: github.com/prometheus/pushgateway
func SplitLabels(labels string, base64JobSuffix string) (map[string]string, error) {
	result := map[string]string{}
	if len(labels) <= 1 {
		// only a single char or nothing? can't be.
		return result, nil
	}

	if strings.HasPrefix(labels, "/") {
		labels = labels[1:]
	}
	components := strings.Split(labels, "/")

	// if odd number of strings: something is wrong with the labels,
	// because they can only occur as key,value pairs.
	if len(components)%2 != 0 {
		return nil, fmt.Errorf("odd number of components in label string %q", labels)
	}

	// loop through every label pair of splitted string.
	for i := 0; i < len(components)-1; i += 2 {
		name, value := components[i], components[i+1]
		trimmedName := strings.TrimSuffix(name, base64JobSuffix)

		// check if label follows Prometheus label rules.
		if !model.LabelNameRE.MatchString(trimmedName) ||
			strings.HasPrefix(trimmedName, model.ReservedLabelPrefix) {
			return nil, fmt.Errorf("improper label name %q", trimmedName)
		}
		if name == trimmedName {
			// value is not base64 encoded, we can continue
			result[name] = value
			continue
		}
		decodedValue, err := DecodeBase64(value)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 encoding for label %s=%q: %v", trimmedName, value, err)
		}
		result[trimmedName] = decodedValue
	}
	return result, nil
}


// GroupingKeyFor creates a grouping key from the provided map of grouping
// labels. The grouping key is created by joining all label names and values
// together with model.SeparatorByte as a separator. The label names are sorted
// lexicographically before joining. In that way, the grouping key is both
// reproducible and unique.
//
// Source: github.com/prometheus/pushgateway
func GroupingKeyFor(labels map[string]string) string {
	if len(labels) == 0 { // Super fast path.
		return ""
	}

	labelNames := make([]string, 0, len(labels))
	for labelName := range labels {
		labelNames = append(labelNames, labelName)
	}
	sort.Strings(labelNames)

	sb := strings.Builder{}
	for i, labelName := range labelNames {
		sb.WriteString(labelName)
		sb.WriteByte(model.SeparatorByte)
		sb.WriteString(labels[labelName])
		if i+1 < len(labels) { // No separator at the end.
			sb.WriteByte(model.SeparatorByte)
		}
	}
	return sb.String()
}

func GroupingKeyForLabelPair(labels []*dto.LabelPair) string {
	m := make(map[string]string)
	for _, label := range labels {
		m[*label.Name] = *label.Value
	}
	return GroupingKeyFor(m)
}

// SanitizeLabels ensures that all the labels in groupingLabels and the
// `instance` label are present in the MetricFamily. The label values from
// groupingLabels are set in each Metric, no matter what. After that, if the
// 'instance' label is not present at all in a Metric, it will be created (with
// an empty string as value).
//
// Finally, sanitizeLabels sorts the label pairs of all metrics.
//
// Source: github.com/prometheus/pushgateway
func SanitizeLabels(mf *dto.MetricFamily, groupingLabels map[string]string) {
	gLabelsNotYetDone := make(map[string]string, len(groupingLabels))

metric:
	for _, m := range mf.GetMetric() {
		for ln, lv := range groupingLabels {
			gLabelsNotYetDone[ln] = lv
		}
		hasInstanceLabel := false
		for _, lp := range m.GetLabel() {
			ln := lp.GetName()
			if lv, ok := gLabelsNotYetDone[ln]; ok {
				lp.Value = proto.String(lv)
				delete(gLabelsNotYetDone, ln)
			}
			if ln == model.InstanceLabel {
				hasInstanceLabel = true
			}
			if len(gLabelsNotYetDone) == 0 && hasInstanceLabel {
				sort.Sort(LabelPairs(m.Label))
				continue metric
			}
		}
		for ln, lv := range gLabelsNotYetDone {
			m.Label = append(m.Label, &dto.LabelPair{
				Name:  proto.String(ln),
				Value: proto.String(lv),
			})
			if ln == model.InstanceLabel {
				hasInstanceLabel = true
			}
			delete(gLabelsNotYetDone, ln) // To prepare map for next metric.
		}
		if !hasInstanceLabel {
			m.Label = append(m.Label, &dto.LabelPair{
				Name:  proto.String(model.InstanceLabel),
				Value: proto.String(""),
			})
		}
		sort.Sort(LabelPairs(m.Label))
	}
}

// TimestampsPresent checks if any timestamps have been specified.
//
// Source: github.com/prometheus/pushgateway
func TimestampsPresent(metricFamilies map[string]*dto.MetricFamily) bool {
	for _, mf := range metricFamilies {
		for _, m := range mf.GetMetric() {
			if m.TimestampMs != nil {
				return true
			}
		}
	}
	return false
}

// LabelPairs implements sort.Interface. It provides a sortable version of a
// slice of dto.LabelPair pointers.
//
// Source: github.com/prometheus/pushgateway
type LabelPairs []*dto.LabelPair

func (s LabelPairs) Len() int {
	return len(s)
}

func (s LabelPairs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s LabelPairs) Less(i, j int) bool {
	return s[i].GetName() < s[j].GetName()
}
