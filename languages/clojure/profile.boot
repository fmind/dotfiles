(merge-env! :dependencies '[[boot-deps "0.1.9"]])

(require '[boot-deps :refer [ancient]])

(task-options!
 repl      {:eval '(set! *print-length* 1000)}
 bare-repl {:eval '(set! *print-length* 1000)})

(deftask with-dw "Add datawalk deps"
  []
  (merge-env! :dependencies '[[datawalk "0.1.12"]])
  (require '[datawalk.core :as dw]))

(deftask cider "Add cider deps."
  []
  (require 'boot.repl)
  (swap! @(resolve 'boot.repl/*default-dependencies*)
         concat '[[org.clojure/tools.nrepl "0.2.12"]
                  [cider/cider-nrepl "0.16.0-SNAPSHOT"]
                  [refactor-nrepl "2.4.0-SNAPSHOT"]])
  (swap! @(resolve 'boot.repl/*default-middleware*)
        concat '[cider.nrepl/cider-middleware
                 refactor-nrepl.middleware/wrap-refactor])
  identity)

(deftask with-io "Add io deps."
  []
  (require '[clojure.java.io :as io]))

(deftask with-sh "Add shell deps."
  []
  (merge-env! :dependencies '[[me.raynes/conch "0.8.0"]])
  (require '[clojure.java.shell :as sh])
  (require '[me.raynes.conch :as ch]))

(deftask with-fs "Add files deps."
  []
  (merge-env! :dependencies '[[me.raynes/fs "1.4.6"]])
  (require '[me.raynes.fs :as fs]))

(deftask with-http "Add http deps."
  []
  (merge-env! :dependencies '[[http-kit "2.2.0"]])
  (require '[org.httpkit.client :as http])
  (import java.net.url))

(deftask with-html "Add enlive deps."
  []
  (merge-env! :dependencies '[[enlive "1.1.6"]])
  (require '[net.cgrand.enlive-html :as html]))

(deftask with-ffox "Add firefox deps."
  []
  (merge-env! :dependencies '[[etaoin "0.1.8-SNAPSHOT"]])
  (require '[etaoin.api :as eta]))

(deftask with-maths "Add maths deps."
  []
  (merge-env! :dependencies '[[org.clojure/math.combinatorics "0.1.4"]
                                     [org.clojure/math.numeric-tower "0.0.4"]])
  (require '[clojure.math.combinatorics :as mc])
  (require '[clojure.math.numeric-tower :as mn]))

(deftask with-stats "Add stats deps."
  []
  (merge-env! :dependencies '[[incanter "1.5.7"]])
  (use '[incanter.core :exclude [trace abs]])
  (require '[incanter.charts :as charts])
  (require '[incanter.bayes :as bayes])
  (require '[incanter.stats :as status])
  (require '[incanter.io :as sio]))

(deftask with-csv "Add csv deps."
  []
  (merge-env! :dependencies '[[clojure-csv/clojure-csv "2.0.2"]])
  (require '[clojure-csv.core :as csv]))

(deftask with-xml "Add xml deps."
  []
  (merge-env! :dependencies '[[org.clojure/data.xml "0.0.8"]])
  (require '[clojure.data.xml :as xml]))

(deftask with-json "Add json deps."
  []
  (merge-env! :dependencies '[[cheshire "5.8.0"]])
  (require '[cheshire.core :as json]))

(deftask with-draw "Add draw deps."
  []
  (merge-env! :dependencies '[[quil "2.6.0"]])
  (require '[quil.core :as ql]))

(deftask with-graph "Add graph deps."
  []
  (merge-env! :dependencies '[[ubergraph "0.4.0"]])
  (require '[ubergraph.core :as g]))

(deftask with-datomic-free "Add datomic free deps."
  []
  (merge-env! :dependencies '[[com.datomic/datomic-free "0.9.5561.62"]])
  (require '[datomic.api :as d]))

(deftask with-datomic-pro "Add datomic pro peer deps."
  []
  (merge-env!
    :repositories [["datomic" {:url "https://my.datomic.com/repo"
                               :username (System/getenv "DATOMIC_REPO_USERNAME")
                               :password (System/getenv "DATOMIC_REPO_PASSWORD")}]]
    :dependencies '[[com.datomic/datomic-pro "0.9.5561.62"]
                    [org.postgresql/postgresql "9.3-1102-jdbc41"]])
  (require '[datomic.api :as d]))
