(merge-env! :dependencies '[[medley "1.0.0"]
                            [spyscope "0.1.5"]
                            [boot-deps "0.1.8"]
                            [adzerk/boot-jar2bin "1.1.0"]
                            [samestep/boot-refresh "0.1.0"]])

(require 'spyscope.core)
(boot.core/load-data-readers!)
(require '[medley.core :refer :all])
(require '[boot-deps :refer [ancient]])
(require '[adzerk.boot-jar2bin :refer [bin]])
(require '[samestep.boot-refresh :refer [refresh]])

(task-options!
 repl      {:eval '(set! *print-length* 1000)}
 bare-repl {:eval '(set! *print-length* 1000)})

(deftask with-io "Add io dependencies." []
  (require '[clojure.java.io :as Io]))

(deftask with-sh "Add shell dependencies." []
  (merge-env! :dependencies '[[me.raynes/conch "0.8.0"]])
  (require '[clojure.java.shell :as Sh])
  (require '[me.raynes.conch :as Ch]))

(deftask with-fs "Add files dependencies." []
  (merge-env! :dependencies '[[me.raynes/fs "1.4.6"]])
  (require '[me.raynes.fs :as Fs]))

(deftask with-http "Add http dependencies." []
  (merge-env! :dependencies '[[http-kit "2.2.0"]])
  (require '[org.httpkit.client :as Http])
  (import java.net.url))

(deftask with-html "Add enlive dependencies." []
  (merge-env! :dependencies '[[enlive "1.1.6"]])
  (require '[net.cgrand.enlive-html :as Html]))

(deftask with-brow "Add browser dependencies." []
  (merge-env! :dependencies '[[etaoin "0.1.8-SNAPSHOT"]])
  (require '[etaoin.api :as Brow]))

(deftask with-maths "Add maths dependencies." []
  (merge-env! :dependencies '[[org.clojure/math.combinatorics "0.1.4"]
                                     [org.clojure/math.numeric-tower "0.0.4"]])
  (use 'clojure.math.combinatorics)
  (use 'clojure.math.numeric-tower))

(deftask with-stats "Add stats dependencies." []
  (merge-env! :dependencies '[[incanter "1.5.7"]])
  (use '[incanter core stats charts]))

(deftask with-csv "Add csv dependencies." []
  (merge-env! :dependencies '[[clojure-csv/clojure-csv "2.0.2"]])
  (require '[clojure-csv.core :as Csv]))

(deftask with-xml "Add xml dependencies." []
  (merge-env! :dependencies '[[org.clojure/data.xml "0.0.8"]])
  (require '[clojure.data.xml :as Xml]))

(deftask with-json "Add json dependencies." []
  (merge-env! :dependencies '[[cheshire "5.8.0"]])
  (require '[cheshire.core :as Json]))

(deftask with-draw "Add draw dependencies." []
  (merge-env! :dependencies '[[quil "2.6.0"]])
  (require '[quil.core :as Draw]))

(deftask with-graph "Add graph dependencies." []
  (merge-env! :dependencies '[[ubergraph "0.4.0"]])
  (require '[ubergraph.core :as Graph]))
