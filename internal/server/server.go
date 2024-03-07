package server

import (
	"net/http"
	"net/http/pprof"

	"github.com/sirupsen/logrus"
)

func healthHandler(w http.ResponseWriter, req *http.Request) {
	_, _ = w.Write([]byte("Ok"))
}

func NewServerMux(log logrus.FieldLogger) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.HandleFunc("/healthz", healthHandler)

	return mux
}
