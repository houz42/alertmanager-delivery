package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/template"
	"github.com/supremind/pkg/duration"
	"github.com/supremind/pkg/log"
)

const defualtTimeout = 1 * time.Minute

var receiverNamePattern = regexp.MustCompile(`^[a-zA-Z0-9-]{1,63}$`)

type Receiver struct {
	client  *http.Client
	tmpl    *template.Template
	conf    *ReceiverConfig
	headers http.Header
	log     logr.Logger
}

type ReceiverConfig struct {
	// name should be a valid dns label, and identical
	Name              string            `json:"name,omitempty"`
	URL               config.URL        `json:"url,omitempty"`
	Body              string            `json:"body,omitempty"`
	AdditionalHeaders map[string]string `json:"additional_headers,omitempty"`
	DownstreamTimeout duration.Duration `json:"downstream_timeout,omitempty"`
}

func NewReceiver(tmpl *template.Template, conf *ReceiverConfig) (*Receiver, error) {
	if !receiverNamePattern.MatchString(conf.Name) {
		return nil, errors.New("no name in Receiver configuration")
	}
	if conf.URL.URL == nil {
		return nil, errors.New("no url in Receiver configuration")
	}
	if tmpl == nil {
		return nil, errors.New("no global template")
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
	body, e := d.transform(r)
	if e != nil {
		return e
	}
	return d.send(ctx, body)
}

func (d Receiver) transform(r io.Reader) (io.Reader, error) {
	if d.conf.Body == "" {
		return r, nil
	}
	data := &template.Data{}
	if e := json.NewDecoder(r).Decode(data); e != nil {
		return nil, fmt.Errorf("decode message: %w", e)
	}

	var e error
	text, e := d.tmpl.ExecuteTextString(d.conf.Body, data)
	if e != nil {
		return nil, fmt.Errorf("execute template: %w", e)
	}
	body := strings.NewReader(text)

	d.log.Info("out", "body", text)
	return body, nil
}

func (d *Receiver) send(ctx context.Context, body io.Reader) error {
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, d.conf.URL.String(), body)
	if e != nil {
		return fmt.Errorf("parse url for new request: %w", e)
	}
	req.Header = d.headers
	resp, e := d.client.Do(req)
	if e != nil {
		return fmt.Errorf("post request: %w", e)
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if l := d.log.V(1); l.Enabled() {
		msg, _ := ioutil.ReadAll(resp.Body)
		l.Info("new message deliveried", "message", body, "response", msg)
	}
	d.log.Info("new message deliveried")

	return nil
}
