package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/template"
	"github.com/supremind/pkg/duration"
	"github.com/supremind/pkg/log"
)

const defualtTimeout = 1 * time.Minute

type Receiver struct {
	client  *http.Client
	tmpl    *template.Template
	conf    *ReceiverConfig
	headers http.Header
	log     logr.Logger
}

type ReceiverConfig struct {
	// name should be a valid path element: no whith space, no slash
	Name              string            `json:"name,omitempty"`
	URL               config.URL        `json:"url,omitempty"`
	Body              string            `json:"body,omitempty"`
	AdditionalHeaders map[string]string `json:"additional_headers,omitempty"`
	DownstreamTimeout duration.Duration `json:"downstream_timeout,omitempty"`
}

func NewReceiver(tmpl *template.Template, conf *ReceiverConfig) (*Receiver, error) {
	if conf.Name == "" {
		return nil, errors.New("no name in Receiver configuration")
	}
	if conf.URL.URL == nil {
		return nil, errors.New("no url in Receiver configuration")
	}
	if conf.Body == "" {
		return nil, errors.New("no body template in Receiver configuration")
	}
	if tmpl == nil {
		return nil, errors.New("empty body template")
	}
	if conf.DownstreamTimeout.Duration <= 0 {
		conf.DownstreamTimeout.Duration = defualtTimeout
	}

	d := &Receiver{
		client: &http.Client{
			Timeout: conf.DownstreamTimeout.Duration,
		},
		tmpl: tmpl,
		conf: conf,
		log:  log.WithName(conf.Name + " receiver"),
	}
	if len(conf.AdditionalHeaders) > 0 {
		h := make(http.Header, len(conf.AdditionalHeaders))
		for k, v := range conf.AdditionalHeaders {
			h.Add(k, v)
		}
		d.headers = h
	}

	return d, nil
}

func (d *Receiver) NewMessage(ctx context.Context, r io.Reader) error {
	data := &template.Data{}
	if e := json.NewDecoder(r).Decode(r); e != nil {
		return fmt.Errorf("decode message: %w", e)
	}

	var body string
	if len(d.conf.Body) > 0 {
		var e error
		body, e = d.tmpl.ExecuteTextString(d.conf.Body, data)
		if e != nil {
			return fmt.Errorf("execute template: %w", e)
		}
	}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, d.conf.URL.String(), strings.NewReader(body))
	if e != nil {
		return fmt.Errorf("parse url for new request: %w", e)
	}
	req.Header = d.headers
	resp, e := d.client.Do(req)
	if e != nil {
		msg, _ := ioutil.ReadAll(resp.Body)
		d.log.Error(e, "request downstream failed", "downstrem response", string(msg))
		return fmt.Errorf("post request: %w", e)
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
	}()

	if l := d.log.V(8); l.Enabled() {
		msg, _ := ioutil.ReadAll(resp.Body)
		l.Info("new message deliveried", "message", body, "response", msg)
	}
	return nil
}
