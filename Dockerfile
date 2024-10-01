FROM golang:1.23 AS builder
WORKDIR /go/src/github.com/sapcc/k8s-conntrack-nanny
ADD . .
RUN CGO_ENABLED=0 go build -v -o /k8s-conntrack-nanny

FROM alpine:3.20
LABEL source_repository="https://github.com/sapcc/k8s-conntrack-nanny"
RUN apk add --no-cache conntrack-tools
COPY --from=builder /k8s-conntrack-nanny /k8s-conntrack-nanny
ENTRYPOINT ["/k8s-conntrack-nanny"]
