(ns user
  (:require [babashka.pods :as pods]
            [clojure.edn :as edn]))

(pods/load-pod 'atomisthq/tools.docker "0.1.0")
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

;; run sbom generation on local image
(docker/sbom "vonwig/clojure-base:jdk17" (fn [event] (println event)))

(docker/hashes "vonwig/malware1:latest" (fn [event] (println event)))

