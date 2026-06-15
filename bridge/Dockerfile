FROM --platform=$BUILDPLATFORM golang:1.26.4-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/lepro-bridge \
    ./cmd/lepro-bridge

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/lepro-bridge /lepro-bridge
USER nonroot:nonroot
ENTRYPOINT ["/lepro-bridge"]
