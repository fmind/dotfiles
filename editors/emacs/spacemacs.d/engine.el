                                        ; IMPORT

(require 'engine-mode)

                                        ; CONFIG

(defengine github "https://github.com/search?q=%s")
(defengine thesaurus "http://thesaurus.com/browse/%s")
(defengine scholar "https://scholar.google.com/scholar?q=%s")
(defengine translate-en "https://translate.google.com/?hl=fr#fr/en/%s")
(defengine translate-fr "https://translate.google.com/?hl=fr#en/fr/%s")

                                        ; BINDINGS

(spacemacs/set-leader-keys "af" 'engine/search-stack-overflow)
(spacemacs/set-leader-keys "ag" 'engine/search-google)
(spacemacs/set-leader-keys "ah" 'engine/search-github)
(spacemacs/set-leader-keys "al" 'engine/search-scholar)
(spacemacs/set-leader-keys "ar" 'engine/search-translate-en)
(spacemacs/set-leader-keys "av" 'engine/search-translate-fr)
(spacemacs/set-leader-keys "aw" 'engine/search-wikipedia)
(spacemacs/set-leader-keys "az" 'engine/search-thesaurus)
