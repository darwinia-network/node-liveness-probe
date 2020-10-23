FROM golang:1.15 as builder

WORKDIR /src
COPY . .

ENV CGO_ENABLED=0

RUN export COMMIT=$(git rev-parse --short HEAD) \
        DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
        TAG=$(git describe --tags --abbrev=0 HEAD) && \
    go build -o node-liveness-probe -ldflags \
        "-X main.buildCommit=$COMMIT \
        -X main.buildDate=$DATE \
        -X main.buildVersion=$TAG" \
    .

FROM alpine

COPY --from=builder /src/node-liveness-probe /usr/bin/node-liveness-probe

EXPOSE 49944
ENTRYPOINT [ "/usr/bin/node-liveness-probe" ]
