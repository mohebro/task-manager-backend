# Build stage: compile a static binary.
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.io,direct
ENV GOSUMDB=off
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Run stage: only the binary, no toolchain.
FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/server ./server
EXPOSE 8080
CMD ["./server"]
