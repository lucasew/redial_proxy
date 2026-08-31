FROM golang:1.25@sha256:698183780de28062f4ef46f82a79ec0ae69d2d22f7b160cf69f71ea8d98bf25d AS build

WORKDIR /app

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . /app

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o app ./cmd/redial_proxy

FROM scratch

COPY --from=build /app/app /app

ENTRYPOINT [ "/app" ]
