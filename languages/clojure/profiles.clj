{
 :user     {:dependencies [[medley "1.0.0"]
                           [alembic "0.3.2"]
                           [org.clojure/tools.namespace "0.2.10"] ]
            :injections [(use '[medley.core])
                         (require '[alembic.still :refer [distill]])
                         (require '[clojure.pprint :refer [pprint]])
                         (require '[clojure.tools.namespace.repl :refer [refresh]])
                         (require '[clojure.repl :refer [apropos dir doc find-doc pst source]])
                         (set! *print-length* 1000)]}
}
