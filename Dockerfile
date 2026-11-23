# Stage 1: Compile the Go application.
FROM golang:1.27-alpine AS build

WORKDIR /src

# Fetch dependencies before copying source so Docker can cache this layer.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /cloud-resource ./cmd/api

# Stage 2: Run only the compiled executable.
FROM scratch
COPY --from=build /cloud-resource /cloud-resource

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/cloud-resource"]
