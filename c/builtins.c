#include "ush.h"

char *builtin_str[] = {
	"cd",
	"help",
	"exit",
	"set",
	"unset",
	"alias",
	"source",
};

char *builtin_help[] = {
	"changes current directory",
	"prints builtin functions",
	"exits the shell",
	"sets environment variable named arg0 to arg1",
	"deletes environment variable named arg0",
	"creates an alias for a command (e.g. `alias ll \"ls -l\")",
	"reads file and executes it's contents",
};

int (*builtin_func[]) (char **) = {
	&ush_cd,
	&ush_help,
	&ush_exit,
	&ush_set,
	&ush_unset,
	&ush_alias,
	&ush_source,
};

int ush_num_builtins() {
	return sizeof(builtin_str) / sizeof(char *);
}

int ush_cd(char **command) {
	char *dir = command[1];
	if (dir == NULL) {
		if (!(dir = getenv("HOME"))) {
			fprintf(stderr, "invalid $HOME\n");
			return 1;
		}
	}
	if (chdir(dir) != 0) {
		fprintf(stderr, "could not change directory to [%s]\n", dir);
		return 1;
	}
	ush_update_cwd();
	return 1;
}

int ush_help(char **command) {
	int i;
	printf("ush builtin functions:\n\n");
	for (i = 0; i < ush_num_builtins(); i++) {
		printf("  %s\t%s\n", builtin_str[i], builtin_help[i]);
	}
	printf("\n");
	return 1;
}

int ush_exit(char **command) {
	return 0;
}

int ush_set(char **command) {
	if (command[1] == NULL || command[2] == NULL) {
		fprintf(stderr, "ush: expected argument name and value to \"set\"\n");
	} else {
		if (setenv(command[1], command[2], 1) != 0) {
			fprintf(stderr, "ush: setenv error\n");
		}
	}
	return 1;
}

int ush_unset(char **command) {
	if (command[1] == NULL) {
		fprintf(stderr, "ush: expected argument to \"unset\"\n");
	} else {
		if (unsetenv(command[1]) != 0) {
			fprintf(stderr, "ush: unsetenv error\n");
		}
	}
	return 1;
}

int ush_alias(char **command) {
	if (command[1] == NULL || command[2] == NULL) {
		fprintf(stderr, "ush: expected 2 arguments to \"alias\"\n");
	} else {
		ush_add_alias(command[1], command[2]);
	}
	return 1;
}

int ush_source(char **command) {
	if (command[1] == NULL) {
		fprintf(stderr, "ush: expected argument to \"source\"\n");
	} else {
		ush_run_file(command[1]);
	}
	return 1;
}
