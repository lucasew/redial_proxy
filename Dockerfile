FROM golang:1.25@sha256:ce63a16e0f7063787ebb4eb28e72d477b00b4726f79874b3205a965ffd797ab2 AS build

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -o app ./cmd/redial_proxy

FROM scratch

COPY --from=build /app/app /app

ENTRYPOINT [ "/app" ]
