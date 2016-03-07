{
 :repl {:dependencies [[org.clojure/tools.nrepl "0.2.12"]]        
        :plugins [[cider/cider-nrepl "0.11.0-SNAPSHOT"]
                  [refactor-nrepl "2.0.0-SNAPSHOT"]]}
 :user {:plugins [[lein-exec "0.3.6"]
                  [lein-ancient "0.6.8"]]
        :dependencies [[slamhound "1.5.5"]]
        :aliases {"slamhound" ["run" "-m" "slam.hound"]}}
}
