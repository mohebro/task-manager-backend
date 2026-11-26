# Build stage: compile a static binary.
# Keep this separate from the deck build: this image is small on
# purpose, because the shared VPS has under 1 GB of RAM and the
# Go build is the heaviest step (add swap if it stalls).
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Run stage: only the binary, no toolchain.
FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/server ./server
EXPOSE 8080
CMD ["./server"]
