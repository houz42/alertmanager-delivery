package receiver

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/template"
)

func TestReceiver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "receiver")
}

var _ = Describe("receiver transformer", func() {
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	var (
		tmpl     *template.Template
		receiver *Receiver
		fakeURL  *config.URL
		// 	downstream *config.URL
		// 	respCh     chan io.ReadCloser
		income string
		data   *template.Data
		now    = time.Now()
	// 	_, _       = respCh, msg
	)

	BeforeEach(func() {
		u, e := url.Parse("fake/url")
		Expect(e).Should(Succeed())
		fakeURL = &config.URL{URL: u}
	})

	Context("when no body template in config is specified", func() {
		BeforeEach(func() {
			receiver = &Receiver{conf: &ReceiverConfig{}}
		})

		It("should pass the content untouched", func() {
			income = "whatever inside"
			out, e := receiver.transform(strings.NewReader(income))
			Expect(e).Should(Succeed())
			Expect(ioutil.ReadAll(out)).Should(Equal([]byte(income)))
		})
	})

	Context("when echo yaml template is loaded globally", func() {

		BeforeEach(func() {
			var e error
			tmpl, e = template.FromGlobs("../example/echo_all.tmpl")
			Expect(e).Should(Succeed())
		})

		Context("if echo yaml template is referenced in body template", func() {
			BeforeEach(func() {
				var e error
				receiver, e = NewReceiver(tmpl, &ReceiverConfig{
					Name: "echo-yaml",
					Body: `{{ template "example.echo-yaml" . }}`,
					URL:  *fakeURL,
				})
				Expect(e).Should(Succeed())

				data = &template.Data{
					Receiver: "deliver-to-echo-yaml",
					Status:   "firing",
					Alerts: []template.Alert{{
						Status:       "firing",
						Labels:       template.KV{"key1": "val1"},
						Annotations:  template.KV{"key2": "val2"},
						StartsAt:     now,
						EndsAt:       now,
						GeneratorURL: "fake/generator",
					}},
					GroupLabels:       template.KV{"key3": "val3"},
					CommonLabels:      template.KV{"key4": "val4"},
					CommonAnnotations: template.KV{"key5": "val5"},
					ExternalURL:       "fake/external",
				}

				jsonIn, e := json.Marshal(data)
				Expect(e).Should(Succeed())
				income = string(jsonIn)
			})

			It("should marshal the content as yaml", func() {
				out, e := receiver.transform(strings.NewReader(income))
				Expect(e).Should(Succeed())

				Expect(ioutil.ReadAll(out)).Should(Equal([]byte(`
receiver: deliver-to-echo-yaml
status: firing
alerts:
- status: firing
  labels:
    key1: val1
  annotations:
    key2: val2
groupLabels:
  key3: val3
commonLabels:
  key4: val4
commonAnnotations:
  key5: val5
`)))
			})

		})
	})
})
