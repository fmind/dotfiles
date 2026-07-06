if status is-interactive
    # a:clear
    abbr -a a clear
    # b:bat
    abbr -a b bat
    # c:gcloud
    abbr -a c gcloud
    abbr -a clog "gcloud auth login --update-adc"
    # d:docker
    abbr -a d docker
    # e:lazydocker
    abbr -a e lazydocker
    # f:fd
    abbr -a f fd
    # g:git
    abbr -a g git
    # go:go
    abbr -a gob "go build"
    abbr -a gom "go mod tidy"
    abbr -a gop pkgsite
    abbr -a gor "go run"
    abbr -a gos "go test ./..."
    abbr -a got "go tool"
    abbr -a gov "govulncheck ./..."
    abbr -a gow "go work"
    abbr -a goz gotestsum
    # h:lazygit
    abbr -a h lazygit
    # i:agy
    abbr -a i agy
    abbr -a iq "agy --prompt"
    # j:just
    abbr -a j just
    # k:kubectl/kubecolor
    if command -q kubecolor
        function kubectl
            kubecolor $argv
        end
        abbr -a k kubecolor
    else
        abbr -a k kubectl
    end
    abbr -a ka "kubectl apply -f"
    abbr -a kd "kubectl describe"
    abbr -a kdel "kubectl delete"
    abbr -a ke "kubectl exec -it"
    abbr -a kga "kubectl get all"
    abbr -a kgd "kubectl get deploy"
    function kget
        command kubectl get $argv -o yaml | kubectl neat
    end
    abbr -a kgp "kubectl get pods"
    abbr -a kgs "kubectl get svc"
    abbr -a kl "kubectl logs"
    abbr -a klf "kubectl logs -f"
    abbr -a kn kubens
    abbr -a kpf "kubectl port-forward"
    abbr -a kx kubectx
    # l:eza
    alias eza="eza --icons=always --git --group-directories-first --time-style=relative --no-quotes"
    alias ls="eza"
    abbr -a l "eza --long --all"
    abbr -a la "eza --all"
    abbr -a ll "eza --long"
    abbr -a lg "eza --long --git --git-ignore"
    abbr -a lt "eza --tree"
    # m:mise
    abbr -a m mise
    abbr -a mr "mise run"
    # n:npm
    abbr -a n npm
    # o:opencode
    abbr -a o opencode
    abbr -a oq "opencode --prompt"
    # p:python
    abbr -a p python3
    abbr -a pt ptpython
    # q:fzf
    abbr -a q fzf
    # r:ripgrep
    abbr -a r rg
    # s:ssh
    abbr -a s ssh
    # t:tofu/terraform
    abbr -a t tofu
    abbr -a tf terraform
    # u:uv
    abbr -a u uv
    abbr -a ur "uv run"
    # v:nvim
    abbr -a v nvim
    abbr -a vi nvim
    # w:zellij
    abbr -a w zellij
    # x:xh
    abbr -a x xh
    # y:yazi
    abbr -a y yazi
end
