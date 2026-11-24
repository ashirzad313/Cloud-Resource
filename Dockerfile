# Stage 1: Compile the Go application.
FROM golang:1.27-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /cloud-resource ./cmd/api

# Stage 2: Run the compiled binary as a non-root user.
FROM scratch
COPY --from=build /cloud-resource /cloud-resource

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/cloud-resource"]
