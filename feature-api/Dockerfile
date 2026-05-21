# ---- Stage 1: Build ------------------------------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY . .
RUN go mod tidy \
 && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /feature-api .

# ---- Stage 2: Runtime ----------------------------------------------
FROM alpine:3.21 AS runtime

RUN apk add --no-cache ca-certificates tzdata \
 && addgroup -g 1001 -S appgroup \
 && adduser  -u 1001 -S appuser -G appgroup

COPY --from=builder /feature-api /usr/local/bin/feature-api

USER appuser

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/feature-api"]
