# syntax=docker/dockerfile:1

FROM golang:1.26.4-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/bankport ./cmd/bankport-api

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/bankport /bankport
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/bankport"]
