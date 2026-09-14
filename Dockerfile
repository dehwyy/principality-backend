FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /principality-web ./cmd/principality-web

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /principality-web /app/principality-web
COPY templates /app/templates
COPY static /app/static
COPY config /app/config
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/principality-web"]
