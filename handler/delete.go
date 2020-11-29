package handler

import (
	"dev.volix.ops/thor/pkg/slog"
	"dev.volix.ops/thor/storage"
	"dev.volix.ops/thor/utils"
	"fmt"
	"github.com/prometheus/common/route"
	"net/http"
	"time"
)

// Delete returns a http.HandlerFunc to delete specific metrics.
// Just like with Push we need a valid job name and labels.
//
// Will return a http.StatusAccepted immediately, as it should
// be clear that the delete action is in any case consistent.
func Delete(ms *storage.MetricStorage, base64 bool) http.HandlerFunc {
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

		ms.SubmitWriteRequest(storage.WriteRequest{
			Labels: labels,
			Timestamp: time.Now(),
		})
		w.WriteHeader(http.StatusAccepted)
	}
}
