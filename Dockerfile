FROM golang:1.23-alpine AS builder
ARG TARGETARCH
COPY influss_linux_${TARGETARCH}/influss /usr/local/bin/
RUN chmod +x /usr/local/bin/influss

FROM alpine:latest
COPY --from=builder /usr/local/bin/influss /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/influss"]
