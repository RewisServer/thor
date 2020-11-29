FROM golang:alpine AS builder

RUN apk add --no-cache git

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

# Download dependencies first, as they don't change as often.
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN sh build.sh

# Multistage to create small image.
FROM scratch

COPY --from=builder /build/thor /

EXPOSE 9091

ENTRYPOINT ["/thor"]

