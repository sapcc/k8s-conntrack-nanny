FROM golang:1.12.14-alpine3.9
WORKDIR /go/src/github.com/sapcc/k8s-conntrack-nanny
ADD . .
RUN go build -v -o /k8s-conntrack-nanny

FROM alpine:3.16
LABEL source_repository="https://github.com/sapcc/k8s-conntrack-nanny"
RUN apk add --no-cache conntrack-tools
COPY --from=0 /k8s-conntrack-nanny /
ENTRYPOINT ["/k8s-conntrack-nanny"]


