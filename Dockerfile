FROM golang:1.20-alpine AS go-builder
RUN set -eux; apk add --no-cache ca-certificates build-base;
RUN apk upgrade --no-cache && apk add bash git make libgcc libc-dev gcc linux-headers eudev-dev jq curl

WORKDIR /go/src/github.com/cosmos/cosmos-sdk
COPY . /go/src/github.com/cosmos/cosmos-sdk/

RUN LEDGER_ENABLED=false BUILD_TAGS=muslc make build


# Final image
FROM alpine:3.15
WORKDIR /root
COPY --from=go-builder /go/src/github.com/cosmos/cosmos-sdk/build/simd /usr/bin/simd
RUN apk upgrade --no-cache && apk add bash git libgcc jq curl tzdata

# Set timezone
ENV TZ Asia/Bangkok

# RUN mkdir -p /root/.six
COPY docker/* /opt/
RUN chmod +x /opt/*.sh

WORKDIR /opt

# Blockchain API
EXPOSE 1317
# Tendermint p2p
EXPOSE 26656
# Tendermint node
EXPOSE 26657

# Run simd by default, omit entrypoint to ease using container with simcli
CMD ["simd"]
