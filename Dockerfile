FROM golang:1.27@sha256:ec47ea3725fea0cc88f6de94919aef19f40f12fc1ec0e2d1b3426dbf15ecca04 AS build

WORKDIR /app

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . /app

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o app ./cmd/redial_proxy

FROM scratch

COPY --from=build /app/app /app

ENTRYPOINT [ "/app" ]
