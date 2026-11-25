FROM golang:1.27@sha256:5f1b98fa02fe396ebce2a0227a0e08309e11792915b9dd2e1f71053d7e7e6666 AS build

WORKDIR /app

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . /app

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o app ./cmd/redial_proxy

FROM scratch

COPY --from=build /app/app /app

ENTRYPOINT [ "/app" ]
