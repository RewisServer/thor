package handler

import (
	"dev.volix.ops/thor/pkg/slog"
	"dev.volix.ops/thor/storage"
	"dev.volix.ops/thor/utils"
	"fmt"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/route"
	"io"
	"mime"
	"net/http"
	"time"
)

const (
	// Suffix for the /job path, so that the job name can be decoded as base64.
	// This is necessary if the name contains `/`, because URLs can not
	// contain slashes.
	Base64JobSuffix = "@base64"
)

// Push returns a http.HandlerFunc to accept pushed metrics to
// the gateway and store them in the storage.MetricStorage.
// If base64 is true, it will try to decode the jobname as base64,
// otherwise not.
// If replace is true, it will remove everything with the grouping key, which is
// just the job name as default, before storing it. Otherwise it will be merged
// with the existing data.
//
// An inconsistent or invalid metric will be rejected with http.StatusBadRequest.
// To skip the slower inconsistency check, unchecked has to be true. This is
// very dangerous though.
//
// Source: github.com/prometheus/pushgateway
func Push(ms *storage.MetricStorage, base64 bool, unchecked bool, replace bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job := route.Param(r.Context(), "job")
		if base64 {
			// we try to decode the job name with base64
			// if it fails, error.
			if job, err := utils.DecodeBase64(job); err != nil {
				http.Error(w, fmt.Sprintf("invalid base64 encoding in job name %q: %v", job, err), http.StatusBadRequest)

				slog.Debug("invalid base64 encoding in job name ", job)
				slog.Debug(err.Error())
				return
			}
		}
		if job == "" {
			http.Error(w, "job name is required", http.StatusBadRequest)

			slog.Debug("job name is required")
			return
		}

		// split labels to get a key,value map for
		// each label.
		labelsString := route.Param(r.Context(), "labels")
		labels, err := utils.SplitLabels(labelsString, Base64JobSuffix)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			slog.Debug("failed to parse url ", labelsString)
			slog.Debug(err.Error())
			return
		}
		labels["job"] = job

		var metricFamilies map[string]*dto.MetricFamily
		ctMediatype, ctParams, ctErr := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if ctErr == nil && ctMediatype == "application/vnd.google.protobuf" &&
			ctParams["encoding"] == "delimited" &&
			ctParams["proto"] == "io.prometheus.client.MetricFamily" {
			// if the body is encoded with protobuf, we can simply
			// decode it and use that.
			metricFamilies = map[string]*dto.MetricFamily{}
			for {
				mf := &dto.MetricFamily{}
				if _, err = pbutil.ReadDelimited(r.Body, mf); err != nil {
					if err == io.EOF {
						err = nil
					}
					break
				}
				metricFamilies[mf.GetName()] = mf
			}
		} else {
			// fallback is a plain/text body.
			var parser expfmt.TextParser
			metricFamilies, err = parser.TextToMetricFamilies(r.Body)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			slog.Debug("failed to parse text from ", r.RemoteAddr)
			slog.Debug(err.Error())
			return
		}

		now := time.Now()
		errCh := make(chan error, 1)

		if unchecked {
			ms.SubmitWriteRequest(storage.WriteRequest{
				Labels:         labels,
				Timestamp:      now,
				MetricFamilies: metricFamilies,
				Replace:        replace,
			})
			w.WriteHeader(http.StatusAccepted)
			return
		}
		// submit write request and consume data which gets send
		// to the Done channel.
		ms.SubmitWriteRequest(storage.WriteRequest{
			Labels:         labels,
			Timestamp:      now,
			MetricFamilies: metricFamilies,
			Replace:        replace,
			Done:           errCh,
		})

		// if an error occurs, we do not want to accept
		// the metric. We only want consistent and valid metrics.
		for err := range errCh {
			http.Error(
				w,
				fmt.Sprintf("pushed metrics are invalid or inconsistent with existing metrics: %v", err),
				http.StatusBadRequest,
			)
			slog.Error(fmt.Sprintf("pushed metrics are invalid or inconsistent with existing metrics (%s, %s): %s",
				r.Method, r.RemoteAddr, err.Error()))
			break
		}
	}
}
