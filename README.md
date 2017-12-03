# μsh

_A shell with a microscopic feature set_

## introduction

`ush` is a simple shell, implementing just the necessary, it currently provides
minimal line editing functions and keyboard shortcuts, simplistic file name
autocompletion, a fixed prompt, piping and a set of 8 builtins.

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
$ echo "/ush/local/bin/ush" | sudo tee -a /etc/shells
$ chsh -s /usr/local/bin/ush
```

## reference

`ush` is pretty minimalistic but still has a lot to offer if you are looking
for just the basics.

**keybindings**

```
enter     execute entered command
ctrl-c    cancel command, start new line
ctrl-d    cancel command, start new line
backspace delete character to the left
up        move to previous command in history
down      move to next command in history
ctrl-u    delete whole line
ctrl-l    clear screen
tab       autocomplete command
```

**builtin commands**

```
help   shows help message
exit   exits the shell
exec   replaces shell with new process
cd     changes current directory
set    sets environment variable named arg1 to arg2
unset  deletes environment variable named arg1
alias  registers a named (arg1) alias for a command (arg2)
source loads and executes a file
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
