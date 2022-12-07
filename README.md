## Background

This is a [babashka pod](https://github.com/babashka/pods) that binds some golang functions into a clojure namespace.  Using this pod, clojure programs can parse dockerfiles and docker images names using the "official" docker golang libraries.  

* [`github.com/docker/distribution/reference`](https://github.com/distribution/distribution/blob/main/reference/reference.go) (for image name parsing)
* [`github.com/moby/buildkit/frontend/dockerfile/parser`](https://github.com/moby/buildkit/blob/master/frontend/dockerfile/parser/parser.go) (for generating a Dockerfile AST).

## Usage

```clojure
(require '[babashka.pods :as pods])
(pods/load-pod 'atomisthq/docker "0.1.0")
; OR use a locally built pod binary
#_(pods/load-pod "./parser")

;; load-pod will create this namespace with two vars
(require '[pod.atomisthq.docker :as docker])

;; parse image names using github.com/docker/distribution 
;; turns golang structs into clojure maps
(docker/parse-image-name "gcr.io/whatever:tag") 
;; automatically turns golang errors into Exceptions
(try
  (docker/parse-image-name "gcr.io/whatever/:tag")
  (catch Exception e 
    ;; invalid reference format
    (println (.getMessage e))))

;; parse dockerfiles using github.com/moby/buildkit
;; returns the Result struct transformed to a clojure map
(docker/parse-dockerfile "FROM \\\n    gcr.io/whatever:tag\nCMD [\"run\"]")
```

Loading `'atomisthq/docker` from the pod registry will download the binary into `${user.home}/.babashka/pods/registry` (the `$BABASHKA_PODS_DIR` environment variable will be used if it exists).

## Building

To build the golang `parser` binary locally, run `go build`.

```bash
go build -o pod-babashka-docker
```

## Contributing

You can find information about contributing to this project in the CONTRIBUTING.md
