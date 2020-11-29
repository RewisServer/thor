package main

import (
	"dev.volix.ops/thor/handler"
	"dev.volix.ops/thor/pkg/slog"
	"dev.volix.ops/thor/pkg/version"
	"dev.volix.ops/thor/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/route"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
)

func main() {
	var (
		app = kingpin.New("thor", "A Prometheus push and aggregation gateway.")

		verbose = app.Flag("verbose", "Enable verbose/debug output.").Default("false").Bool()

		listenAddress        = app.Flag("web.listen-address", "Address and port to listen on.").Default(":9091").String()
		metricsPath          = app.Flag("web.metrics-path", "Path under which to expose metrics.").Default("/metrics").String()
		skipConsistencyCheck = app.Flag("push.skip-consistency-check", "Skip consistency check, dangerous but faster.").Default("false").Bool()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *verbose {
		// we only support verbose or !verbose, as we don't need
		// a more specific setting like debug, info, warn, ... level.
		slog.SetVerbosity(1)
	}

	slog.Info("starting thor gateway version ", version.Version)
	slog.Info("build context: ", version.BuildContext())

	slog.Debug("listen address=", *listenAddress)
	slog.Debug("metrics path=", *metricsPath)

	ms := storage.NewMetricStorage()

	r := route.New()
	r.Get("/-/healthy", handler.Health(ms))
	r.Get("/lore", handler.Lore())

	// POST merges and adds to it and PUT replaces
	for _, suffix := range []string{"", handler.Base64JobSuffix} {
		isBase64 := suffix == handler.Base64JobSuffix

		r.Post(*metricsPath+"/job"+suffix+"/:job/*labels", handler.Push(ms, isBase64, *skipConsistencyCheck, false))
		r.Put(*metricsPath+"/job"+suffix+"/:job/*labels", handler.Push(ms, isBase64, *skipConsistencyCheck, true))
		r.Del(*metricsPath+"/job"+suffix+"/:job/*labels", handler.Delete(ms, isBase64))

		r.Post(*metricsPath+"/job"+suffix+"/:job", handler.Push(ms, isBase64, *skipConsistencyCheck,false))
		r.Put(*metricsPath+"/job"+suffix+"/:job", handler.Push(ms, isBase64, *skipConsistencyCheck,true))
		r.Del(*metricsPath+"/job"+suffix+"/:job", handler.Delete(ms, isBase64))
	}

	// create gatherer to serve /metrics page
	g := prometheus.Gatherers{
		prometheus.GathererFunc(func() ([]*dto.MetricFamily, error) { return ms.GetMetricFamilies(), nil }),
	}
	r.Get(*metricsPath, promhttp.HandlerFor(g, promhttp.HandlerOpts{}).ServeHTTP)

	mux := http.NewServeMux()
	mux.Handle("/", r)

	server := &http.Server{
		Addr:    *listenAddress,
		Handler: mux,
	}
	err := server.ListenAndServe()
	slog.Error("http server stopped: ", err)
}
