package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/alertmanager/template"
)

type Config struct {
	Address string `json:"address,omitempty"`
	// Templates *
	Deliveries []*DeliveryConfig `json:"deliveries,omitempty"`
}

func Serve(ctx context.Context, conf *Config) error {
	mux := http.NewServeMux()
	// todo
	var tmpl *template.Template

	for _, cf := range conf.Deliveries {
		d, e := NewDelivery(tmpl, cf)
		if e != nil {
			return fmt.Errorf("config delivery handler: %w", e)
		}
		mux.Handle(cf.Name, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			e := d.NewMessage(ctx, r.Body)
			if e != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				// todo: log
				rw.Write([]byte(e.Error()))
				return
			}
			rw.WriteHeader(http.StatusOK)
		}))

		// todo: log new handler
	}

	return nil
}
