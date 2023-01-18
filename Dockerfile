FROM golang:1.19-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY main.go ./
COPY docker/ ./docker/
COPY babashka/ ./babashka/

RUN CGO_ENABLED=0 go build -o pod-atomisthq-tools.docker

FROM alpine:3.17

COPY repository/ /root/.babashka/pods/repository
COPY --from=build /app/pod-atomisthq-tools.docker /root/.babashka/pods/repository/atomisthq/tools.docker/0.1.0
RUN chmod 755 /root/.babashka/pods/repository/atomisthq/tools.docker/0.1.0/pod-atomisthq-tools.docker
