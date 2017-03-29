{
 :user {:plugins [[lein-try "0.4.3"]
                  [lein-exec "0.3.6"]
                  [lein-cljfmt "0.5.6"]
                  [lein-ancient "0.6.10"]
                  ;;[venantius/ultra "0.5.1"]
                  [io.aviso/pretty "0.1.26"]
                  [jonase/eastwood "0.2.3"]]
        :dependencies [[alembic "0.3.2"]
                       [spyscope "0.1.6"]
                       [io.aviso/pretty "0.1.26"]
                       [im.chit/vinyasa "0.4.7"]
                       [org.clojure/tools.nrepl "0.2.12"]
                       [org.clojure/tools.namespace "0.2.10"]
                       [leiningen #=(leiningen.core.main/leiningen-version)]]
        :injections [(require 'clojure.repl)
                     (require 'spyscope.core)
                     (require '[vinyasa.inject :as inject])
                     (inject/in
                      [alembic.still :refer [distill]]
                      [clojure.pprint :refer [pprint]]
                      [clojure.java.shell :refer [sh]]
                      [clojure.tools.namespace.repl :refer [refresh]]
                      [clojure.repl :refer [apropos dir doc source find-doc]]
                      ;;[vinyasa.lein :exclude [*project*]]
                      ;;[vinyasa.inject :refer [inject [in inject-in]]]
                      [vinyasa.reflection .> .? .* .% .%> .& .>ns .>var]
                      )]}
 :sh [:user
         {:dependencies [[me.raynes/fs "1.4.6"]]
          :injections[(use '[clojure.java.shell])
                      (require '[me.raynes.fs :as fs])]}]
 :http [:user
        {:dependencies [[http-kit "2.2.0"]
                        [enlive "1.1.6"]]
         :injections [(require '[org.httpkit.client :as http])
                      (require '[net.cgrand.enlive-html :as live])
                      (import java.net.URL)]}]
 :selenium [:user
            {:dependencies [[clj-webdriver/clj-webdriver "0.7.2"]
                           [org.seleniumhq.selenium/selenium-java "3.0.1"]]
            :injections [(use 'clj-webdriver.taxi)
                         (set-driver! {:browser :firefox})]}]
 :maths [:user
         {:dependencies [[org.clojure/math.combinatorics "0.1.4"]
                         [org.clojure/math.numeric-tower "0.0.4"]]
          :injections [(use 'clojure.math.combinatorics)
                       (use 'clojure.math.numeric-tower)]}]
 :stats [:user
         {:dependencies [[incanter "1.9.1"]]
          :injections [(use '[incanter core stats charts])]}]

 :data [:user
        {:dependencies [[cheshire "5.7.0"]
                        [org.clojure/data.xml "0.0.8"]
                        [clojure-csv/clojure-csv "2.0.2"]]
         :injections [(require '[cheshire.core :as json])
                      (require '[clojure.data.xml :as xml])
                      (require '[clojure-csv.core :as csv])]}]
 :graph [:user
         {:dependencies [[aysylu/loom "0.6.0"]]
          :injections [(require '[loom
                                   [graph :as Graph]
                                   [attr :as attr]
                                   [alg :as Alg]
                                   [gen :as Gen]
                                   [io :as io]])]}]
 :nb [:user
      {:dependencies [[lein-gorilla "0.4.0"]]
       :injections [(require '[gorilla-repl.core :refer [run-gorilla-server]])
                    (run-gorilla-server{:port 8990})
                    (./sh "xdg-open" "http://127.0.0.1:8990/worksheet.html")]}]
 :draw [:user
         {:dependencies [[quil "2.6.0"]]
          :injections [(require '[quil.core :as draw])]}]
}
