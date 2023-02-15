FROM golang:1.19-alpine AS build

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
COPY --from=build /app/babashka-pod-docker /root/.babashka/pods/repository/docker/babashka-pod-docker/0.1.0
RUN chmod 755 /root/.babashka/pods/repository/docker/babashka-pod-docker/0.1.0/babashka-pod-docker
