# Stage 1: Compile the Go application.
FROM golang:1.27-alpine AS build

WORKDIR /src
COPY . .

# A static Linux binary does not need the Go toolchain at runtime.
RUN CGO_ENABLED=0 GOOS=linux go build -o /cloud-resource ./cmd/api

# Stage 2: Run only the compiled executable.
FROM scratch

COPY --from=build /cloud-resource /cloud-resource

# Run the server as a non-root user.
USER 65532:65532

# Document the port used by our application.
EXPOSE 8080

ENTRYPOINT ["/cloud-resource"]
