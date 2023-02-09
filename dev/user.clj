(ns user
  (:require [babashka.pods :as pods]
            [clojure.edn :as edn]
            [babashka.pods.impl :as impl]))

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

(comment
  (pods/load-pod 'docker/babashka-pod-docker "0.1.0")

  (require '[babashka-pod-docker :as docker])


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
  )

(defn generate-sbom
  [image]
  (impl/invoke-public
   "docker.babashka-pod-docker"
   "babashka-pod-docker/generate-sbom"
   [image "" ""]
   {:handlers {:done (fn [] (println "Done"))
               :success (fn [msg] (println "msg: " msg))
               :error (fn [_err] #_"TODO: handle this error")}}))

(comment
  (println (load-pod "./babashka-pod-docker"))
  (impl/invoke-public
   "docker.babashka-pod-docker"
   "babashka-pod-docker/generate-sbom"
   ["ubuntu:latest" "" ""]
   {})
  (generate-sbom "alpine")
  )
