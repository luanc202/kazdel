FROM golang:1.25.6-alpine AS build
WORKDIR /usr/src/app/go/api
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o url-shortener ./api/

FROM alpine:3.23.3 AS api
WORKDIR /usr/src/app/go/api
COPY --from=build /usr/src/app/go/api/url-shortener .
RUN chmod +x url-shortener
EXPOSE 3000
CMD ["./url-shortener"]
