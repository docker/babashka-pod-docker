(ns main
  (:require [babashka.pods :as pods]
            [clojure.edn :as edn]
            [babashka.curl :as curl]
            [clojure.string :as string]))

(def x (pods/load-pod 'atomisthq/tools.docker "0.1.0"))
(require '[pod.atomisthq.docker :as docker])

(defn do-transaction [all-hashes transactions m token digest]
  (let [tx-data (->> @all-hashes
                     (filter (fn [{:keys [path]}] (if path (string/includes? path ".exe"))))
                     (mapcat (fn [{:keys [hash diff-id]}]
                               (let [blob-digest (get m diff-id)]
                                 (if blob-digest
                                   [{:schema/entity blob-digest
                                     :schema/entity-type :docker.image/blob
                                     :docker.image.blob/digest blob-digest}
                                    {:schema/entity-type :docker.image.blob/file
                                     :docker.image.blob.file/sha256 hash
                                     :docker.image.blob.file/blob blob-digest}]
                                   (do
                                     (println diff-id "not in " m)
                                     [])))))
                     (into []))]
    (try
      (println "tx-data" tx-data)
      (println
       (curl/post transactions
                  {:body (pr-str {:transactions [{:data tx-data}]})
                   :headers {"Authorization" (format "Bearer %s" token)
                             "Content-Type" "application/edn"}}))
      (println
       (curl/post transactions
                  {:body (pr-str {:transactions [{:data [{:docker.image/digest digest
                                                          :schema/entity-type :docker/image
                                                          :malware.status/indexed :malware.status.indexed/complete}]}]})
                   :headers {"Authorization" (format "Bearer %s" token)
                             "Content-Type" "application/edn"}}))
      (System/exit 0)
      (catch Throwable t
        (println "error " t)
        (System/exit 1)))))

(defn transact-hashes [{:keys [image digest m transactions token]}]
  (println image digest transactions)
  (let [all-hashes (atom [])]
    (docker/hashes image (fn [event]
                           (if (= "done" (:status event))
                             (do-transaction all-hashes transactions m token digest)
                             (swap! all-hashes conj (edn/read-string event)))))))

#_(let [[image digest m transaction-url token] *command-line-args*]
    (transact-hashes {:image image :digest digest :diff-id->digest (edn/read-string m) :transaction-url transaction-url :token token}))

(transact-hashes (edn/read-string (slurp "/Users/slim/atmhq/malware/test1.edn")))
(while true (Thread/sleep 5000))
