package storage

import (
	"dev.volix.ops/thor/pkg/slog"
	"dev.volix.ops/thor/utils"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"sync"
	"time"
)

// A MetricGroup is a wrapper for a map of metric families.
// A group is unique by its set of labels (namely grouping keys).
//
// But note that this is not an idea taken by Prometheus in general,
// it's only used for something like this gateway, as we have a better way
// to group multiple metrics together with similar labels.
//
// A good example would be labels like `instance` and `job`.
//
// It is recommended to not carry a pointer to this struct around,
// as it would only be more efficient, if the Labels would be a large
// map, but it is rather small. And there could be consistency issues
// if the Labels map change during runtime by accident.
type MetricGroup struct {
	Labels         map[string]string
	MetricFamilies map[string]*dto.MetricFamily
}

// A MetricStorage is the in-memory storage of all metrics pushed
// to this gateway.
//
// Note that this way every Metric get lost, if the gateway restarts.
// But in general this is not a problem at all with Prometheus, because
// of its good data consistency checks and fetch logic.
//
// Take a Counter for example: if the counter is 0 for a small period
// of time and then gets back up, all Prometheus queries still just work
// fine - just like before.
//
// All actions done with this storage are secured by a sync.RWMutex to enable
// concurrent access to the groups.
// Also a writeQueue is available to push WriteRequest into a channel
// instead of synchronous data storage.
type MetricStorage struct {
	lock         sync.RWMutex
	writeQueue   chan WriteRequest
	metricGroups map[string]MetricGroup
}

// A request to write the containing MetricFamilies to
// the MetricStorage.
//
// If MetricFamilies is nil - not empty! - we count that as a delete request.
//
// If Replace is true, already existing data similar to
// the family, i.e. metrics with the same names, will be deleted.
//
// If Done is nil, this request will not trigger the expensive
// consistency check.
type WriteRequest struct {
	Labels         map[string]string
	Timestamp      time.Time
	MetricFamilies map[string]*dto.MetricFamily
	Replace        bool
	Done           chan error
}

const (
	// How many requests we allow in the queue at the same time.
	// Every request exceeding this limit will be discarded.
	writeQueueCapacity = 1000
)

func NewMetricStorage() *MetricStorage {
	ms := &MetricStorage{
		writeQueue:   make(chan WriteRequest, writeQueueCapacity),
		metricGroups: make(map[string]MetricGroup),
	}

	go ms.loop()
	return ms
}

func NewSimpleMetricStorage() *MetricStorage {
	ms := &MetricStorage{
		metricGroups: make(map[string]MetricGroup),
	}
	return ms
}

func (ms *MetricStorage) SubmitWriteRequest(wr WriteRequest) {
	ms.writeQueue <- wr
}

// Same as GetMetricGroups but it resolves the groups
// and returns a slice of all metric families.
func (ms *MetricStorage) GetMetricFamilies() []*dto.MetricFamily {
	ms.lock.RLock()
	defer ms.lock.RUnlock()

	var result []*dto.MetricFamily
	for _, group := range ms.metricGroups {
		for _, family := range group.MetricFamilies {
			result = append(result, utils.CopyMetricFamily(family))
		}
	}
	return result
}

// GetMetricGroups returns a copy of all current
// MetricGroup
func (ms *MetricStorage) GetMetricGroups() map[string]MetricGroup {
	ms.lock.RLock()
	defer ms.lock.RUnlock()

	groupsCopy := make(map[string]MetricGroup, len(ms.metricGroups))
	for k, g := range ms.metricGroups {
		metricsCopy := make(map[string]*dto.MetricFamily, len(g.MetricFamilies))
		for n, mf := range g.MetricFamilies {
			metricsCopy[n] = utils.CopyMetricFamily(mf)
		}

		groupsCopy[k] = MetricGroup{Labels: g.Labels, MetricFamilies: metricsCopy}
	}
	return groupsCopy
}

// Healthy returns if the storage is in a deadlock
// or if the request queue is too full.
// Returns <nil> if everything is good.
func (ms *MetricStorage) Healthy() error {
	// check for deadlocks
	ms.lock.Lock()
	defer ms.lock.Unlock()

	// check if write queue is available for new requests
	if ms.writeQueue != nil && len(ms.writeQueue) >= cap(ms.writeQueue) {
		return fmt.Errorf("write queue is full")
	}
	return nil
}

// loop loops through the write queue of the
// MetricStorage and checks for new requests.
func (ms *MetricStorage) loop() {
	for {
		select {
		case wr := <-ms.writeQueue:
			// we do simple consistency checks.
			// if the done channel of wr is existent, we suppose
			// that we want to do the heavy check as well.
			var err error
			if err = validateConsistency(ms, wr); err == nil {
				ms.processWriteRequest(wr)
			} else {
				wr.Done <- err
			}

			if wr.Done != nil {
				close(wr.Done)
			}
		}
	}
}

// processWriteRequest takes the WriteRequest and stores them in the MetricStorage.
// If no previous group exist or if WriteRequest.Replace is true, then
// we simply put the MetricGroup into the storage.
// Otherwise we merge the two groups and their metrics together.
func (ms *MetricStorage) processWriteRequest(wr WriteRequest) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	groupingKey := utils.GroupingKeyFor(wr.Labels)

	if wr.MetricFamilies == nil {
		// if no metric families are given, the body has
		// to be empty. So we delete everything with this groupingKey.
		delete(ms.metricGroups, groupingKey)
		return
	}

	group := MetricGroup{
		Labels:         wr.Labels,
		MetricFamilies: wr.MetricFamilies,
	}

	prevGroup, ok := ms.metricGroups[groupingKey]
	if !ok || wr.Replace {
		// either group does not exist, we can just create a new one
		// and we're done.
		// or we want to replace the whole group, and we're done too.
		ms.metricGroups[groupingKey] = group
		return
	}
	// if not, we merge the groups
	mergeGroups(prevGroup, group)
}

// validateConsistency return if applying the provided WriteRequest will result in
// a consistent state of metrics. The dms is not modified by the check. However,
// the WriteRequest _will_ be sanitized: the MetricFamilies are ensured to
// contain the grouping Labels after the check. If false is returned, the
// causing error is written to the Done channel of the WriteRequest.
//
// Special case: If the WriteRequest has no Done channel set, the (expensive)
// consistency check is skipped. The WriteRequest is still sanitized, and the
// presence of timestamps still results in returning false.
//
// Source: github.com/prometheus/pushgateway
func validateConsistency(ms *MetricStorage, wr WriteRequest) error {
	if wr.MetricFamilies == nil {
		// Delete request cannot create inconsistencies, and nothing has
		// to be sanitized.
		return nil
	}

	// check for duplicates, but with different types
	// as we can't merge them anyway.
	for _, f2 := range wr.MetricFamilies {
		for _, group := range ms.metricGroups {
			f1, ok := group.MetricFamilies[*f2.Name]
			if ok && *f1.Type != *f2.Type {
				return fmt.Errorf("cannot merge metric '%s': type %s != %s", *f1.Name, f1.Type.String(), f2.Type.String())
			}
		}
	}

	if utils.TimestampsPresent(wr.MetricFamilies) {
		return fmt.Errorf("pushed metrics must not have timestamps")
	}
	for _, mf := range wr.MetricFamilies {
		utils.SanitizeLabels(mf, wr.Labels)
	}

	// Without Done channel, don't do the expensive consistency check.
	if wr.Done == nil {
		return nil
	}

	// Construct a test metric storage, acting on a copy of the metrics, to test the
	// WriteRequest with.
	testMs := &MetricStorage{
		metricGroups: ms.GetMetricGroups(),
	}
	testMs.processWriteRequest(wr)

	// Construct a test Gatherer to check if consistent gathering is possible.
	tg := prometheus.Gatherers{
		prometheus.GathererFunc(func() ([]*dto.MetricFamily, error) {
			return testMs.GetMetricFamilies(), nil
		}),
	}
	if _, err := tg.Gather(); err != nil {
		return err
	}
	return nil
}

// mergeGroups takes two MetricGroup and merge their families
// together.
// For that it checks if the name of the family is the same.
//   1. If not, just add family to group.
//   2. If, merge the families with mergeFamilies together.
// g1 is now the merged group.
func mergeGroups(g1, g2 MetricGroup) {
	for key, g2Family := range g2.MetricFamilies {
		g1Family, ok := g1.MetricFamilies[key]
		if !ok {
			// element does not exist, we put it into the map.
			g1.MetricFamilies[key] = g2Family
			continue
		}

		// element does exist, merge family
		err := mergeFamilies(g1Family, g2Family)
		if err != nil {
			// if we cannot merge the metric, we just skip it
			slog.Debug(err.Error())
		}
	}
}

// mergeFamilies merges the second family into the first one,
// such that the first one then has all new information in it.
//
// We know beforehand that both groupingKeys of the families are the same
// therefore we only need to take all specific metrics which the family contains
// and push them onto the first family.
// The algorithm works as follows:
//   1. Create a unique key for every metric in f1
//   2. Loop through every metric in f2
//   3. Compare key of metric with f1 metrics
//     3a. Key does not exist? Put it into f1
//     3b. Key does exist? Merge content of metrics
//   4. f1 is now the family with the updated metrics
//
// Returns an error if e.g. the family types are not equal.
func mergeFamilies(f1, f2 *dto.MetricFamily) error {
	if *f1.Type != *f2.Type {
		// if types are not equal, we can cancel immediately
		return fmt.Errorf("cannot merge metric '%s': type %s != %s", *f1.Name, f1.Type.String(), f2.Type.String())
	}

	// we map a metric grouping key to its metric
	// so that we can search faster for duplicates when
	// comparing with metrics of family f2.
	mm := make(map[string]*dto.Metric)
	for _, metric := range f1.Metric {
		key := utils.GroupingKeyForLabelPair(metric.Label)

		mm[key] = metric
	}

	for _, f2Metric := range f2.Metric {
		key := utils.GroupingKeyForLabelPair(f2Metric.Label)

		f1Metric, ok := mm[key]
		if !ok {
			// metric does not exist, so we add it to the list
			f1.Metric = append(f1.Metric, f2Metric)
		} else {
			// otherwise, we merge the metrics together
			mergeMetrics(*f1.Type, f1Metric, f2Metric)
		}
	}
	return nil
}

// mergeMetrics takes two metrics of the same type and
// combines them together.
//
// We simply use switch-case for every single metric type.
func mergeMetrics(mt dto.MetricType, m1, m2 *dto.Metric) {
	switch mt {
	case dto.MetricType_COUNTER:
		*m1.Counter.Value += *m2.Counter.Value
	case dto.MetricType_GAUGE:
		// there is no reason to add gauges together.
		// that's we, we just SET the value and we're done.
		*m1.Gauge.Value = *m2.Gauge.Value
	case dto.MetricType_HISTOGRAM:
		hist1, hist2 := m1.Histogram, m2.Histogram

		*hist1.SampleCount += *hist2.SampleCount
		*hist1.SampleSum += *hist2.SampleSum
		mergeBuckets(&hist1.Bucket, &hist2.Bucket)
	case dto.MetricType_SUMMARY:
		// impossible to merge, as the calculation for
		// the quantile values expect a specific algorithm
		// which we should not reimplement ourselves.

		// we just override the old one.
		*m1.Summary = *m2.Summary
	case dto.MetricType_UNTYPED:
		// here as well: no reason for us to add them together.
		// just setting the value is enough
		*m1.Untyped.Value = *m2.Untyped.Value
	}
}

// mergeBuckets combines two slices of buckets together.
// For that we just have to add up the internal cumulative
// count for the specific buckets.
// Reason for that is that the upper bound of the buckets dont
// change during the runtime, so two buckets with the same bound
// can simply be added together, without destroying the metric.
func mergeBuckets(b1, b2 *[]*dto.Bucket) {
	// create a map for faster access
	// with its upper bound as key
	mm := make(map[float64]*dto.Bucket)
	for _, bucket := range *b1 {
		mm[*bucket.UpperBound] = bucket
	}

	for _, b2Bucket := range *b2 {
		b1Bucket, ok := mm[*b2Bucket.UpperBound]
		if !ok {
			// buckets are different, just add it
			*b1 = append(*b1, b2Bucket)
		} else {
			// buckets are the same, we merge these two
			*b1Bucket.CumulativeCount += *b2Bucket.CumulativeCount
		}
	}
}
