{
 :user     {:plugins [[lein-codox "0.10.3"]
                      [lein-ancient "0.6.10"]
                      [jonase/eastwood "0.2.4"]]
            :dependencies [[medley "1.0.0"]
                           [alembic "0.3.2"]
                           [org.clojure/tools.namespace "0.2.10"] ]
             :injections [(use '[medley.core])
                          (require '[alembic.still :refer [distill]])
                          (require '[clojure.pprint :refer [pprint]])
                          (require '[clojure.tools.namespace.repl :refer [refresh]])
                          (require '[clojure.repl :refer [apropos dir doc find-doc pst source]])
                          (set! *print-length* 1000)]}
 :nb       [:user
            {:dependencies [[lein-gorilla "0.4.0"]]
             :injections   [(require '[gorilla-repl.core :refer [run-gorilla-server]])
                            (run-gorilla-server {:port 8990})
                            (./sh "xdg-open" "http://127.0.0.1:8990/worksheet.html")]}]
 :sh       [:user
            {:dependencies [[me.raynes/fs "1.4.6"]]
             :injections   [(use '[clojure.java.shell])
                            (require '[me.raynes.fs :as Fs])]}]
 :http     [:user
            {:dependencies [[http-kit "2.2.0"]
                            [enlive "1.1.6"]]
             :injections   [(require '[org.httpkit.client :as Http])
                            (use '[net.cgrand.enlive-html])
                            (import java.net.URL)]}]
 :selenium [:user
            {:dependencies [[clj-webdriver/clj-webdriver "0.7.2"]
                            [org.seleniumhq.selenium/selenium-java "3.5.3"]]
             :injections   [(use 'clj-webdriver.taxi)
                            (set-driver! {:browser :firefox})]}]
 :maths    [:user
            {:dependencies [[org.clojure/math.combinatorics "0.1.4"]
                            [org.clojure/math.numeric-tower "0.0.4"]]
             :injections   [(use 'clojure.math.combinatorics)
                            (use 'clojure.math.numeric-tower)]}]
 :stats    [:user
            {:dependencies [[incanter "1.9.1"]]
             :injections   [(use '[incanter core stats charts])]}]

 :data     [:user
            {:dependencies [[cheshire "5.8.0"]
                            [org.clojure/data.xml "0.0.8"]
                            [clojure-csv/clojure-csv "2.0.2"]]
             :injections   [(require '[clojure.java.io :as Io])
                            (require '[cheshire.core :as Json])
                            (require '[clojure.data.xml :as Xml])
                            (require '[clojure-csv.core :as Csv])]}]
 :graph    [:user
            {:dependencies [[aysylu/loom "1.0.0"]]
             :injections   [(require '[loom
                                       [graph :as Graph]
                                       [attr :as Attr]
                                       [alg :as Alg]
                                       [gen :as Gen]
                                       [io :as Io]])]}]
 :draw     [:user
            {:dependencies [[quil "2.6.0"]]
             :injections   [(require '[quil.core :as Draw])]}]
 }
