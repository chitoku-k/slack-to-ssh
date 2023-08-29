# syntax = docker/dockerfile:1
FROM golang:1.21.0 AS build
WORKDIR /usr/src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go \
    go mod download
COPY . ./
RUN --mount=type=cache,target=/go \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags='-s -w'

FROM scratch
ARG PORT=443
ENV PORT=$PORT
ENV GIN_MODE=release
COPY --link --from=build /lib/x86_64-linux-gnu/ld-linux-x86-64.* /lib/x86_64-linux-gnu/
COPY --link --from=build /lib/x86_64-linux-gnu/libc.so* /lib/x86_64-linux-gnu/
COPY --link --from=build /lib/x86_64-linux-gnu/libresolv.so* /lib/x86_64-linux-gnu/
COPY --link --from=build /lib64 /lib64
COPY --link --from=build /usr/src/slack-to-ssh /slack-to-ssh
COPY --link --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE $PORT
CMD ["/slack-to-ssh"]
