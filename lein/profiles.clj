{
 :repl {:dependencies [[org.clojure/tools.nrepl "0.2.12"]]        
        :plugins [[cider/cider-nrepl "0.11.0-SNAPSHOT"]
                  [refactor-nrepl "2.0.0-SNAPSHOT"]]}
 :user {:plugins [[lein-exec "0.3.6"]
                  [lein-ancient "0.6.8"]
                  [lein-cljfmt "0.4.1"]
                  [lein-cloverage "1.0.6"]
                  [jonase/eastwood "0.2.3"]
                  [lein-kibit "0.1.2"]
                  [lein-bikeshed "0.3.0"]
                  [venantius/ultra "0.4.1"]]
        :dependencies [[slamhound "1.5.5"]]
        :aliases {"slamhound" ["run" "-m" "slam.hound"]}}
}
