FROM golang:1.19-alpine AS build

RUN apk --no-cache add git openssh-client

ENV GOPRIVATE=github.com/docker
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN --mount=type=secret,id=gitghatoken \
    (test -f /run/secrets/gitghatoken && \
    git config --global url."https://x-access-token:$(cat /run/secrets/gitghatoken)@github.com/docker".insteadOf "https://github.com/docker" || \
    git config --global url."git@github.com:docker/".insteadOf "https://github.com/docker/")

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY main.go ./
COPY docker/ ./docker/
COPY babashka/ ./babashka/

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o babashka-pod-docker

FROM alpine:3.17
ARG version
COPY repository/ /root/.babashka/pods/repository
COPY --from=build /app/babashka-pod-docker /root/.babashka/pods/repository/docker/docker-tools/0.1.0
RUN chmod 755 /root/.babashka/pods/repository/docker/docker-tools/0.1.0/babashka-pod-docker
