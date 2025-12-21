# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod ./
# no go.sum yet; keep simple for scaffold
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 go build -o /out/agentos ./cmd/agentos

FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl
COPY --from=build /out/agentos /usr/local/bin/agentos
ENTRYPOINT ["agentos"]
