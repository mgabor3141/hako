# Build the fixture MCP server into a tiny static image (no shell needed — the
# mock runs no external commands; the callHook lives in the gateway, not here).
FROM golang:1.25.5 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /github-mock .

FROM gcr.io/distroless/static-debian12
COPY --from=build /github-mock /github-mock
EXPOSE 8080
ENTRYPOINT ["/github-mock"]
