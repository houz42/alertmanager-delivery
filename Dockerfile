FROM golang:1.14-alpine3.12 as builder

ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /src
COPY . /src
RUN go build -o _output/alertmanager-delivery .

FROM alpine:3.12
COPY --from=builder /src/_output/alertmanager-delivery /usr/bin/alertmanager-delivery
COPY example /example
ENTRYPOINT [ "/usr/bin/alertmanager-delivery" ]
