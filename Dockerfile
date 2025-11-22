FROM golang:1.25@sha256:f60eaa87c79e604967c84d18fd3b151b3ee3f033bcdade4f3494e38411e60963 AS build

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -o app ./cmd/redial_proxy

FROM scratch

COPY --from=build /app/app /app

ENTRYPOINT [ "/app" ]
