{
  :repl {:global-vars {*print-length* 1000}}

  :user {:dependencies []
         :plugins [[lein-ancient "0.6.15"]]
         ; :global-vars {*warn-on-reflection* true}
         :injections [(require '[clojure.repl :refer :all])
                      (require '[clojure.pprint :refer :all])]}
}
