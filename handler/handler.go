package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/alertmanager/template"
	"github.com/supremind/pkg/log"
)

type Config struct {
	// path to global template files
	Templates []string          `json:"templates,omitempty"`
	Receivers []*ReceiverConfig `json:"receivers,omitempty"`
}

func Serve(ctx context.Context, conf *Config) (*http.ServeMux, error) {
	tmpl, e := template.FromGlobs(conf.Templates...)
	if e != nil {
		return nil, fmt.Errorf("parse global templates: %w", e)
	}

	mux := http.NewServeMux()
	for _, cf := range conf.Receivers {
		d, e := NewReceiver(tmpl, cf)
		if e != nil {
			return nil, fmt.Errorf("config delivery handler %s: %w", cf.Name, e)
		}

		l := log.WithName(cf.Name + " handler")
		mux.Handle("/"+cf.Name, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			l.Info("new message")
			e := d.NewMessage(ctx, r.Body)
			if e != nil {
				l.Error(e, "send message to downstream")
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(e.Error()))
				return
			}
			rw.WriteHeader(http.StatusOK)
		}))

		log.WithName("server").Info("receiver handler registered", "receiver name", cf.Name)
	}

	return mux, nil
}
