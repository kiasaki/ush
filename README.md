# μsh

_A shell with a microscopic feature set_

## introduction

`ush` is a simple shell, built atop the great **lineoise** line editing library, implementing just the necessary. It currently implements nice line editing, simple wordexp based autocompletion, a fixed prompt and the help, cd and exit builtins.

## installing

```
make clean install
```

## using

You can either start `ush` from your current shell like so:

```
$ ush
```

or set it as your default login shell using

```
$ chsh -s /usr/local/bin/ush
```

## reference

`ush` is pretty minimalistic but still has a lot to offer if you are looking for just the basics.

**keybindings**

```
enter     execute entered command
ctrl-c    cancel command, start new line
backspace delete charater to the left
ctrl-d    delete charater to the right
ctrl-t    swap current character with previous
ctrl-b    move cursor left
ctrl-f    move cursor right
ctrl-p    move to previous command in history
ctrl-n    move to next command in history
ctrl-u    delete whole line
ctrl-k    delete from current to the end of the line
ctrl-a    goto the start of the line
ctrl-e    goto the end of the line
ctrl-l    clear screen
ctrl-w    delete previous word
tab       autocomplete command
```

**builtin commands**

```
help      shows the help message
cd        changes the current directory
exit      quits the shell session
```

## missing

**Missing aliases?**

You can get aliases by create a shell script in a `bin/` directory that is in you path:

`~/bin/ll`

```sh
#!/bin/sh
ls -la
```

**Missing a fancy colored prompt?**

Prompts, like color schemes and theme, because the plethora of choices can be a great waste
of time. Try using ush's prompt a little bit, you'll get used to it pretty fast.

Now, you might be missing information that used to be displayed for you in that prompt you had.
Here are a few commands that can give you exactly that info only when you actually need it:

```
ush$ pwd
/Users/kiasaki/code/repos/ush
ush$ date
Sun 16 Oct 2016 11:46:26 EDT
ush$ whoami
kiasaki
ush$ git status -sb
## master
```

## license

MIT. See `LICENSE` file.

Lineoise has it's own license at the top of the `lineoise.c` file.

