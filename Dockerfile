FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder
WORKDIR /ingestion
# Copy local code to the container image.

COPY  .  ./

ARG TARGETOS
ARG TARGETARCH

RUN go mod download


RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -mod=readonly -v -o ingestion-server ./cmd/main.go

FROM alpine:3.23

RUN addgroup -S web && \
    adduser -S -D -h /home/webuser -s /bin/sh -G web webuser && \
    mkdir -p /home/webuser && \
    chown -R webuser:web /home/webuser

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

USER webuser
# needed for the temp file to be written to a directory with right permission
WORKDIR /home/webuser
# Copy the binary to the production image from the builder stage.
COPY --from=builder --chown=root:root  --chmod=755 /ingestion/ingestion-server /home/webuser/ingestion-server

# Run the web service on container startup.
CMD ["/home/webuser/ingestion-server", "rest-server"]