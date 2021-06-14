function my_github_clone_https -d "clone a https repository from my github account."
    if ! set --query argv[1]
        echo "Missing argv[1]: repository" >&2
        return 1
    end
    set --local repository $argv[1]
    git clone "https://fmind@github.com/fmind/$repository.git"
end
