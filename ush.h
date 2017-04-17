#ifndef _USH_H
#define _USH_H

#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include <string.h>
#include <unistd.h>
#include <limits.h>
#include <wordexp.h>
#include <sys/wait.h>
#include <sys/stat.h>
#include "linenoise.h"

#define USH_VERSION "0.0.1"

#define EXIT_SUCCESS 0
#define EXIT_FAILURE 1

#define MAX_PROMPT_SZ 1024

// Init
void ush_update_prompt(void);

void ush_run_file(char *filename);

// Builtins
int ush_cd(char **command);
int ush_help(char **command);
int ush_exit(char **command);
int ush_setenv(char **command);
int ush_unsetenv(char **command);
int ush_source(char **command);

extern char *builtin_str[6];
extern char *builtin_help[6];
extern int (*builtin_func[6]) (char **);

int ush_num_builtins(void);

// Parsing
char **ush_parse(char *line);

#endif
