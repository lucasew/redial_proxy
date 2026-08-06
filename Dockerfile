FROM golang:1.26@sha256:2005724102f45917a63e9d092fc0e4ea56ea575048ce147caad5f5f61502c365 AS build

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -o app ./cmd/redial_proxy

FROM scratch

COPY --from=build /app/app /app

ENTRYPOINT [ "/app" ]
