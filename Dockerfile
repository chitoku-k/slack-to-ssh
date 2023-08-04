# syntax = docker/dockerfile:1
FROM golang:1.20.7 AS build
WORKDIR /usr/src
COPY go.mod go.sum /usr/src/
RUN --mount=type=cache,target=/go \
    go mod download
COPY . /usr/src/
RUN --mount=type=cache,target=/go \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags='-s -w'

FROM scratch
ARG PORT=443
ENV PORT=$PORT
ENV GIN_MODE=release
COPY --from=build /usr/src/slack-to-ssh /slack-to-ssh
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE $PORT
CMD ["/slack-to-ssh"]
