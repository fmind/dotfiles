import snake
import re


@snake.key_map("<leader>H")
def toggle_snake_case_camel_case():
    word = snake.get_word()

    if "_" in word:
        chunks = word.split("_")
        camel_case = chunks[0] + "".join([chunk.capitalize() for chunk in chunks[1:]])
        snake.replace_word(camel_case)
    else:
        chunks = filter(None, re.split("([A-Z][^A-Z]*)", word))
        snake_case = "_".join([chunk.lower() for chunk in chunks])
        snake.replace_word(snake_case)
