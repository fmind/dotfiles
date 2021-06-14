function my_github_clone_ssh -d "clone a ssh repository from my github account."
    if ! set --query argv[1]
        echo "Missing argv[1]: repository" >&2
        return 1
    end
    set --local repository $argv[1]
    git clone "git@github.com:fmind/$repository.git"
end
