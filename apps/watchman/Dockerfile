FROM golang:1.26.5-alpine AS watchman-build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/watchman ./cmd/watchman

FROM golang:1.26.5-alpine AS goose-build

RUN go install -tags="no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_ydb no_vertica" \
    github.com/pressly/goose/v3/cmd/goose@v3.27.1

FROM alpine:3.24

RUN apk add --no-cache ca-certificates && addgroup -S atalaya && adduser -S atalaya -G atalaya
WORKDIR /app
COPY --from=watchman-build /out/watchman /usr/local/bin/watchman
COPY --from=goose-build /go/bin/goose /usr/local/bin/goose
COPY migrations /app/migrations
USER atalaya
EXPOSE 8080
CMD ["watchman"]

