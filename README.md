## Background

This is a [babashka pod](https://github.com/babashka/pods) that binds some golang functions into a clojure namespace.  Using this pod, clojure programs can parse dockerfiles and docker images names using the "official" docker golang libraries.

* [`github.com/docker/distribution/reference`](https://github.com/distribution/distribution/blob/main/reference/reference.go) (for image name parsing)
* [`github.com/moby/buildkit/frontend/dockerfile/parser`](https://github.com/moby/buildkit/blob/master/frontend/dockerfile/parser/parser.go) (for generating a Dockerfile AST).

## Usage

```clojure
(require '[babashka.pods :as pods])
(pods/load-pod 'docker/tools "0.1.0")
; OR use a locally built pod binary
#_(pods/load-pod "./babashka-pod-docker")

;; load-pod will create this namespace with two vars
(require '[docker.tools :as docker])

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

Loading `'docker/docker-tools` from the pod registry will download the binary into `${user.home}/.babashka/pods/registry` (the `$BABASHKA_PODS_DIR` environment variable will be used if it exists).

## Building Locally

To build the golang `parser` binary locally, run `go build`.

```bash
go build -o babashka-pod-docker
```

## Releasing

All pushes to main will update the 0.1.0 release. This is becaus maintaining the pod version in the repository directory and in the pod registry is tricky.

We hope to automate all of that in the future.

## Namespace generation

The `pods/load-pod` call is convenient for a repl-session, or a script, but what if you are `aot` compiling, or building a native binary.  In the example above, the namespaces emitted by `pods/load-pod` are not available until runtime.

Here is an example of bindings that will resolve at compile-time and go through the same dispatch.

```clj
; require the babashka.pods in a namespace
(require '[babashka.pods.impl :as impl])

; call at runtime to initialize pod system
(defn load-pod
  ([pod-spec] (load-pod pod-spec nil))
  ([pod-spec version opts] (load-pod pod-spec (assoc opts :version version)))
  ([pod-spec opts]
   (let [opts (if (string? opts)
                {:version opts}
                opts)
         pod (impl/load-pod
              pod-spec
              (merge {:remove-ns remove-ns
                      :resolve (fn [sym]
                                 (or (resolve sym)
                                     (intern
                                      (create-ns (symbol (namespace sym)))
                                      (symbol (name sym)))))}
                     opts))]
     (future (impl/processor pod))
     {:pod/id (:pod-id pod)})))

;; statically define dispatch functions - this is synchronous
(defn parse [s]
  (impl/invoke-public "docker.tools" "docker.tools/parse-dockerfile" [s] {}))

;; async example
(defn generate-sbom [s]
  (impl/invoke-public "docker.tools" "docker.tools/generate-sbom"
    [s cb]
    {:handlers {:done (fn [])
                :success cb
                :error (fn [err]}})))
```

```
(pods/load-pod "/bin/babashka-pod-docker")
```

This method of dispatch does not require any dynamic namespace generation.

## Contributing

You can find information about contributing to this project in the CONTRIBUTING.md

