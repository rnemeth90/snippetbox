FROM golang:1.23-alpine AS build

RUN apk add --no-cache build-base
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /snippetbox ./cmd/web

FROM alpine:3.21

RUN apk add --no-cache ca-certificates libgcc \
    && adduser -D -H -u 10001 app
COPY --from=build /snippetbox /usr/local/bin/snippetbox

USER 10001
EXPOSE 4000
ENTRYPOINT ["/usr/local/bin/snippetbox"]
