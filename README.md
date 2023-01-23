## Background

This is a [babashka pod](https://github.com/babashka/pods) that binds some golang functions into a clojure namespace.  Using this pod, clojure programs can parse dockerfiles and docker images names using the "official" docker golang libraries.  

* [`github.com/docker/distribution/reference`](https://github.com/distribution/distribution/blob/main/reference/reference.go) (for image name parsing)
* [`github.com/moby/buildkit/frontend/dockerfile/parser`](https://github.com/moby/buildkit/blob/master/frontend/dockerfile/parser/parser.go) (for generating a Dockerfile AST).

## Usage

```clojure
(require '[babashka.pods :as pods])
(pods/load-pod 'atomisthq/tools.docker "0.1.0")
; OR use a locally built pod binary
#_(pods/load-pod "./pod-atomisthq-tools.docker")

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

Create `vonwig/pod-atomisthq-tools.docker` which is a manifest list with pod binaries for both `amd64` and `arm64`.  This image is a good way to pull the pod binaries into skill containers.

```bash
bb build-pod-image
```

## Namespace generation

The `pods/load-pod` call is convenient for a repl-session, or a script, but what if you are `aot` compiling, or building a native binary.  In the example above, the namespaces emitted by `pods/load-pod` are not available until runtime.

Here is an example of bindings that will resolve at compile-time and go through the same dispatch.

```
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

;; statically define dispatch functions
(defn parse [s]
  (impl/invoke-public "pod.atomisthq.docker" "pod.atomisthq.docker/parse-dockerfile" [s] {}))
```

```
(pods/load-pod 'atomisthq/tools.docker "7.3.0")
(pods/load-pod "my-executable")
```

This method of dispatch does not require any dynamic namespace generation.

## Contributing

You can find information about contributing to this project in the CONTRIBUTING.md
