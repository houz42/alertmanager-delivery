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

	"github.com/prometheus/alertmanager/template"
)

type Delivery struct {
	client  *http.Client
	tmpl    *template.Template
	conf    *DeliveryConfig
	headers http.Header
}

func NewDelivery(tmpl *template.Template, conf *DeliveryConfig) (*Delivery, error) {
	if conf.Name == "" {
		return nil, errors.New("no name in delivery configuration")
	}
	if conf.URL == "" {
		return nil, errors.New("no url in delivery configuration")
	}
	if conf.Template == "" {
		return nil, errors.New("no template in delivery configuration")
	}
	if tmpl == nil {
		return nil, errors.New("empty template")
	}

	d := &Delivery{
		// todo: timeout
		client: http.DefaultClient,
		tmpl:   tmpl,
		conf:   conf,
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

type DeliveryConfig struct {
	Name              string            `json:"name,omitempty"`
	URL               string            `json:"url,omitempty"`
	Template          string            `json:"template,omitempty"`
	AdditionalHeaders map[string]string `json:"additional_headers,omitempty"`
}

func (d *Delivery) NewMessage(ctx context.Context, r io.Reader) error {
	data := &template.Data{}
	if e := json.NewDecoder(r).Decode(r); e != nil {
		return fmt.Errorf("decode message: %w", e)
	}

	var body string
	if len(d.conf.Template) > 0 {
		var e error
		body, e = d.tmpl.ExecuteTextString(d.conf.Template, data)
		if e != nil {
			return fmt.Errorf("execute template: %w", e)
		}
	}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, d.conf.URL, strings.NewReader(body))
	if e != nil {
		return fmt.Errorf("parse url for new request: %w", e)
	}
	req.Header = d.headers
	resp, e := d.client.Do(req)
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
	}()
	if e != nil {
		return fmt.Errorf("post request: %w", e)
	}

	// todo: log resp
	return nil
}
