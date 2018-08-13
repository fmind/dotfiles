import snake


@snake.key_map("<leader>zb")
def toggle_boolean():
    word = snake.get_word()

    if word == "True":
        snake.replace_word("False")
    elif word == "False":
        snake.replace_word("True")
