FROM golang:1.15 as build-env

WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...

RUN go build -o /go/bin/app

FROM gcr.io/distroless/base:nonroot
LABEL org.opencontainers.image.source https://github.com/coralogix/prometheus-alert-readiness
COPY --from=build-env /go/bin/app /
USER nonroot
CMD ["/app"]
