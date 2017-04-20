# μsh

_A shell with a microscopic feature set_

## introduction

`ush` is a simple shell, built atop the great **liner** line editing library (similar to **linenoise**), implementing just the necessary. It currently implements nice line editing, simple file name autocompletion, a fixed prompt, piping and the cd, set, unset, alias and exit builtins.

## installing

```
make
make install
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
cd     changes current directory
exit   exits the shell
set    sets environment variable named arg1 to arg2
unset  deletes environment variable named arg1
alias  registers a named (arg1) alias for a command (arg2)
```

## missing

**Missing redirection?**

For input (`<`) try using `cat` and piping that into the command.

For output (`>`) try using `tee` and pipes.

For more complex command try building them in multiple steps, using intermediary
files or creating a script file for it.

**Missing a fancy colored prompt?**

Prompts, like color schemes and theme, are not a thing in `ush` because the plethora
of choice and constant switching and tweaking can be a great waste of time. Try using
ush's prompt a little bit, you'll get used to it pretty fast.

Now, you might be missing information that used to be displayed for you in that prompt you had.
Here are a few commands that can give you exactly that info only when you actually need it:

```
ush$ pwd
/Users/kiasaki/code/repos/ush
ush$ date
Sun 16 Oct 2016 11:46:26 EDT
ush$ whoami
kiasaki
ush$ hostname
kiasaki-mbp
ush$ git status -sb
## master
```

## license

MIT. See `LICENSE` file.
