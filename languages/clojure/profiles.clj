{
  :repl {:global-vars {*print-length* 1000}}

  :user {:dependencies [[spyscope "0.1.6"]]
         :plugins [[lein-ancient "0.6.15"]]
         ; :global-vars {*warn-on-reflection* true}
         :injections [(require 'spyscope.core)
                      (require '[clojure.repl :refer :all])
                      (require '[clojure.pprint :refer :all])]}
}
