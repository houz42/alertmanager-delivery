[![Actions Status](https://github.com/houz42/alertmanager-delivery/workflows/Alertmanager%20Delivery/badge.svg)](https://github.com/houz42/alertmanager-delivery/actions?query=workflow%3A%22Alertmanager+Delivery%22)
[![Coverage Status](https://coveralls.io/repos/github/houz42/alertmanager-delivery/badge.svg?branch=master)](https://coveralls.io/github/houz42/alertmanager-delivery?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/houz42/alertmanager-delivery)](https://goreportcard.com/report/github.com/houz42/alertmanager-delivery) 


# Alertmanager Delivery

The missing template rendering for Alertmanager webhook receivers.

The Prometheus developers are "[not actively adding new receivers](https://prometheus.io/docs/alerting/latest/configuration/#receiver)" for Alertmanager, and recommend implementing custom notification integrations via the webhook receiver. 
This project helps connecting Alertmanger with any webhook servers, in any format you want. 
The Delivery works between Alertmanager and downstream webhook servers, transforms messages from the Alertmanger as defined in the templates, and deliveries them to the downstream.


## Run an example server
### start an echo server as the downstream
```
$ docker run -p 5678:5678 --rm hashicorp/http-echo -text="hello world"
```

### start the delivery server
```
$ go run main.go --config example/config.yaml
```

### send message to delivery server as the Alertmanager will do
```
$ curl localhost:41357/echo-yaml -X POST -d @example/message.json
```

## Deployment
[TODO]

## Configuration
[TODO]

## Included Templates

- [Wechat work group bot](https://work.weixin.qq.com/help?doc_id=13376)
