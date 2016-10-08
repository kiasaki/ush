#ifndef _USH_H
#define _USH_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>
#include <unistd.h>
#include "linenoise.h"

#define USH_VERSION "0.0.1"

#define EXIT_SUCCESS 0
#define EXIT_FAILURE 1

// Builtins
int ush_cd(char **command);
int ush_help(char **command);
int ush_exit(char **command);

extern char *builtin_str[3];
extern char *builtin_help[3];
extern int (*builtin_func[3]) (char **);

int ush_num_builtins(void);

// Parsing
#define USH_TOKEN_BUFER_SIZE 64
#define USH_TOKEN_DELIMITER " \t\r\n\a"

#endif
