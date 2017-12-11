(merge-env! :dependencies '[[medley "1.0.0"]
                            ; [spyscope "0.1.5"]
                            [boot-deps "0.1.8"]
                            ; [datawalk "0.1.4-SNAPSHOT"]
                            [adzerk/boot-jar2bin "1.1.0"] ])

; (require 'spyscope.core)
; (boot.core/load-data-readers!)
(require '[medley.core :refer :all])
(require '[boot-deps :refer [ancient]])
; (require '[datawalk.core :as Datawalk])
(require '[adzerk.boot-jar2bin :refer [bin]])

(task-options!
 repl      {:eval '(set! *print-length* 1000)}
 bare-repl {:eval '(set! *print-length* 1000)})

(deftask with-io "Add io deps."
  []
  (require '[clojure.java.io :as Io]))

(deftask with-sh "Add shell deps."
  []
  (merge-env! :dependencies '[[me.raynes/conch "0.8.0"]])
  (require '[clojure.java.shell :as Sh])
  (require '[me.raynes.conch :as Ch]))

(deftask with-fs "Add files deps."
  []
  (merge-env! :dependencies '[[me.raynes/fs "1.4.6"]])
  (require '[me.raynes.fs :as Fs]))

(deftask with-http "Add http deps."
  []
  (merge-env! :dependencies '[[http-kit "2.2.0"]])
  (require '[org.httpkit.client :as Http])
  (import java.net.url))

(deftask with-html "Add enlive deps."
  []
  (merge-env! :dependencies '[[enlive "1.1.6"]])
  (require '[net.cgrand.enlive-html :as Html]))

(deftask with-ffox "Add firefox deps."
  []
  (merge-env! :dependencies '[[etaoin "0.1.8-SNAPSHOT"]])
  (require '[etaoin.api :as Brow]))

(deftask with-maths "Add maths deps."
  []
  (merge-env! :dependencies '[[org.clojure/math.combinatorics "0.1.4"]
                                     [org.clojure/math.numeric-tower "0.0.4"]])
  (use 'clojure.math.combinatorics)
  (use 'clojure.math.numeric-tower))

(deftask with-stats "Add stats deps."
  []
  (merge-env! :dependencies '[[incanter "1.5.7"]])
  (require '[incanter.charts :as Charts])
  (require '[incanter.bayes :as Bayes])
  (require '[incanter.stats :as Stats])
  (require '[incanter.io :as Statsio])
  (use '[incanter.core :exclude [trace abs]])
  )

(deftask with-csv "Add csv deps."
  []
  (merge-env! :dependencies '[[clojure-csv/clojure-csv "2.0.2"]])
  (require '[clojure-csv.core :as Csv]))

(deftask with-xml "Add xml deps."
  []
  (merge-env! :dependencies '[[org.clojure/data.xml "0.0.8"]])
  (require '[clojure.data.xml :as Xml]))

(deftask with-json "Add json deps."
  []
  (merge-env! :dependencies '[[cheshire "5.8.0"]])
  (require '[cheshire.core :as Json]))

(deftask with-draw "Add draw deps."
  []
  (merge-env! :dependencies '[[quil "2.6.0"]])
  (require '[quil.core :as Draw]))

(deftask with-music "Add music deps."
  []
  (merge-env! :dependencies '[[overtone "0.10.3"]])
  #_(use 'overtone.core))

(deftask with-graph "Add graph deps."
  []
  (merge-env! :dependencies '[[ubergraph "0.4.0"]])
  (require '[ubergraph.core :as Graph]))

(deftask with-gremlin "Add gremlin deps."
  []
  (merge-env! :dependencies '[[clojurewerkz/ogre "3.3.0.0"]
                              [org.apache.tinkerpop/tinkergraph-gremlin "3.3.0"]])
  (import '[org.apache.tinkerpop.gremlin.tinkergraph.structure TinkerGraph TinkerFactory])
  (require '[clojurewerkz.ogre.core :as Grem]))

(deftask with-cassandra "Add cass deps."
  []
  (merge-env! :dependencies '[[cc.qbits/alia-all "4.0.3"]])
  (require '[qbits.alia :as Alia]))

(deftask with-datomic-pro "Add datomic pro peer deps."
  []
  (merge-env!
    :repositories [["datomic" {:url "https://my.datomic.com/repo"
                               :username (System/getenv "DATOMIC_REPO_USERNAME")
                               :password (System/getenv "DATOMIC_REPO_PASSWORD")}]]
    :dependencies '[[com.datomic/datomic-pro "0.9.5561.62"]
                    [com.datastax.cassandra/cassandra-driver-core "3.1.0"] ])
  (require '[datomic.api :as Dt]))

(deftask with-datomic-free "Add datomic free peer deps."
  []
  (merge-env! :dependencies '[[com.datomic/datomic-free "0.9.5561.62"]])
  (require '[datomic.api :as Dt]))

(deftask with-datomic-client "Add datomic client deps."
  []
  (merge-env! :dependencies '[[com.datomic/clj-client "0.8.606"]])
  (require '[datomic.client :as Dt]))
