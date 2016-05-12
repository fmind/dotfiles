{
 :user {:plugins [[lein-exec "0.3.6"]
                  [lein-cljfmt "0.5.3"]
                  [lein-ancient "0.6.10"]]
        :dependencies [[alembic "0.3.2"]
                       [spyscope "0.1.5"]
                       [im.chit/vinyasa "0.4.3"]
                       [io.aviso/pretty "0.1.26"]
                       [org.clojure/tools.nrepl "0.2.12"]
                       [org.clojure/tools.namespace "0.2.10"]
                       [leiningen #=(leiningen.core.main/leiningen-version)]]
        :injections [(require 'spyscope.core)
                     (require 'io.aviso.repl)
                     (require '[vinyasa.inject :as inject])
                     (inject/in
                      [vinyasa.inject :refer [inject [in inject-in]]]
                      [vinyasa.lein :exclude [*project*]]
                      [alembic.still [distill pull]]
                      clojure.core [vinyasa.reflection .> .? .* .% .%> .& .>ns .>var]
                      clojure.core > [clojure.pprint pprint] [clojure.java.shell sh])]
        }

 :http [:user
        {:dependencies [[http-kit "2.1.19"]
                        [enlive "1.1.6"]]
         :injections [(require '[org.httpkit.client :as http])
                      (use 'net.cgrand.enlive-html)
                      (import java.net.URL)]}]

 :selenium [:user
            {:dependencies [[clj-webdriver/clj-webdriver "0.7.2"]
                           [org.seleniumhq.selenium/selenium-java "2.52.0"]]
            :injections [(use 'clj-webdriver.taxi)
                         (set-driver! {:browser :firefox})]}]
 :maths [:user
         {:dependencies [[org.clojure/math.combinatorics "0.1.1"]
                         [org.clojure/math.numeric-tower "0.0.4"]]
          :injections [(use 'clojure.math.combinatorics)
                       (use 'clojure.math.numeric-tower)]}]

 :stats [:user
         {:dependencies [[incanter "1.5.7"]]
          :injections [(use '[incanter core stats charts])]}]

 :data [:user
        {:dependencies [[cheshire "5.6.1"]
                        [org.clojure/data.xml "0.0.8"]
                        [clojure-csv/clojure-csv "2.0.1"]]
         :injections [(use 'cheshire.core)
                      (use 'clojure.data.xml)
                      (use 'clojure-csv.core)]}]
 :graph [:user
         {:dependencies [[aysylu/loom "0.6.0"]]
          :injections [(use '[loom graph alg gen attr label io derived])]}]
}
