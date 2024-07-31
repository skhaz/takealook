# syntax=docker/dockerfile:1

FROM golang:1.22 AS modules
WORKDIR /modules
COPY go.mod go.sum ./
RUN go mod download

FROM golang:1.22 AS builder
COPY --from=modules /go/pkg /go/pkg
WORKDIR /opt
COPY . .
RUN <<EOF
    VERSION=$(grep -oE "playwright-go v\S+" /opt/go.mod | sed 's/playwright-go //g')
    go install github.com/playwright-community/playwright-go/cmd/playwright@${VERSION}
    go build -ldflags="-s -w" -trimpath -o app
EOF

FROM debian:bookworm-slim AS extensions
WORKDIR /opt
RUN <<EOF
    apt-get update
    apt-get install --no-install-recommends --yes ca-certificates curl jq unzip

    IFS=$'\n\t'

    REPO_URL="https://api.github.com/repos/uBlockOrigin/uBOL-home/releases/latest"
    ASSET_URL=$(curl -sSL "$REPO_URL" | jq -r '.assets[] | select(.name | endswith(".chromium.mv3.zip")).browser_download_url')
    FILENAME=$(basename "$ASSET_URL")

    curl -sSL -o "$FILENAME" "$ASSET_URL" && unzip "$FILENAME" -d ublock

    REPO_URL="https://api.github.com/repos/OhMyGuus/I-Still-Dont-Care-About-Cookies/releases/latest"
    ASSET_URL=$(curl -sSL "$REPO_URL" | jq -r '.assets[] | select(.name | endswith("-chrome-source.zip")).browser_download_url')
    FILENAME=$(basename "$ASSET_URL")

    curl -sSL -o "$FILENAME" "$ASSET_URL" && unzip "$FILENAME" -d isdncac
EOF

FROM debian:bookworm-slim
WORKDIR /usr/local/bin
RUN <<EOF
    apt-get update
    apt-get install --no-install-recommends --yes \
    fonts-crosextra-carlito \
    fonts-crosextra-caladea \
    fonts-dejavu \
    fonts-droid-fallback \
    fonts-freefont-ttf \
    fonts-liberation \
    fonts-liberation2 \
    fonts-noto \
    fonts-noto-cjk \
    fonts-noto-color-emoji \
    fonts-opensymbol \
    fonts-roboto \
    fonts-roboto-unhinted \
    fonts-sil-gentium \
    fonts-sil-gentium-basic \
    fonts-stix \
    fonts-symbola \
    ttf-bitstream-vera \
    ca-certificates \
    tzdata \
    imagemagick
    rm -rf /var/lib/apt/lists/*
EOF
COPY --from=builder /go/bin/playwright .
RUN playwright install --with-deps chromium
WORKDIR /opt/extensions
COPY --from=extensions /opt/ublock/ ublock
COPY --from=extensions /opt/isdncac isdncac
WORKDIR /opt
COPY --from=builder /opt/app .
CMD ["/opt/app"]
