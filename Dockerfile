# syntax = docker/dockerfile:1.4
FROM nixos/nix:latest AS builder

WORKDIR /tmp/build
RUN mkdir /tmp/nix-store-closure

RUN \
    --mount=type=cache,target=/nix,from=nixos/nix:latest,source=/nix \
    --mount=type=cache,target=/root/.cache \
    --mount=type=bind,target=/tmp/build \
    <<EOF
  ls -l /nix/store | wc
  nix \
    --extra-experimental-features "nix-command flakes" \
    --extra-substituters "http://host.docker.internal?priority=10" \
    --option filter-syscalls false \
    --show-trace \
    --log-format raw \
    build . --out-link /tmp/output/result
  cp -R $(nix-store -qR /tmp/output/result) /tmp/nix-store-closure
EOF

FROM scratch as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY main.go ./
COPY docker/ ./docker/
COPY babashka/ ./babashka/

ENV GOPRIVATE=github.com/docker
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN --mount=type=secret,id=gitghatoken \
    (test -f /run/secrets/gitghatoken && \
    git config --global url."https://x-access-token:$(cat /run/secrets/gitghatoken)@github.com/docker".insteadOf "https://github.com/docker" || \
    git config --global url."git@github.com:docker/".insteadOf "https://github.com/docker/")

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o babashka-pod-docker

FROM alpine:3.17
ARG version
COPY repository/ /root/.babashka/pods/repository
COPY --from=build /app/babashka-pod-docker /root/.babashka/pods/repository/docker/docker-tools/0.1.0
RUN chmod 755 /root/.babashka/pods/repository/docker/docker-tools/0.1.0/babashka-pod-docker
