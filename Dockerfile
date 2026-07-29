# Build stage
FROM golang:1.25.8-alpine AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
# Note: we removed GOARCH=amd64 so it correctly uses the host's architecture
RUN CGO_ENABLED=0 GOOS=linux go build -v -o app ./cmd/web

# Final stage
FROM alpine:3.19

WORKDIR /usr/src/app

COPY --from=builder /usr/src/app/pkg/ui/static ./pkg/ui/static
COPY --from=builder /usr/src/app/app ./

# Add execution permission just to be safe
RUN chmod +x ./app
RUN mkdir -p data

EXPOSE 3000
CMD ["./app"]
