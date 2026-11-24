# Stage 1: Compile Go applications.
FROM golang:1.27-alpine AS build

WORKDIR /src

# Cache external dependencies while source files change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /cloud-resource ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -o /vmcheck ./cmd/vmcheck

# Stage 2: Run compiled binaries without a shell or Go compiler.
FROM scratch
COPY --from=build /cloud-resource /cloud-resource
# Temporary repository inspection tool; remove when the VM API is in place.
COPY --from=build /vmcheck /vmcheck

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/cloud-resource"]
