FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

RUN go install github.com/google/wire/cmd/wire@latest

# Copy the rest of the source tree
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s -extldflags '-static'" \
    -trimpath \
    -o /out/server \
    ./cmd/main.go


FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/server /server

# 1323 is the API port.  pprof (6060) stays internal — never map it publicly.
EXPOSE 1323

# Run as the built-in nonroot user (uid 65532)
USER nonroot:nonroot

ENTRYPOINT ["/server"]
