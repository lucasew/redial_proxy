FROM golang:1.26@sha256:3aff6657219a4d9c14e27fb1d8976c49c29fddb70ba835014f477e1c70636647 AS build

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -o app ./cmd/redial_proxy

FROM scratch

COPY --from=build /app/app /app

ENTRYPOINT [ "/app" ]
