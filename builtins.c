#include "ush.h"

char *builtin_str[] = {
	"cd",
	"help",
	"exit",
	"setenv",
	"unsetenv",
	"source",
};

char *builtin_help[] = {
	"changes current directory",
	"prints builtin functions",
	"exits the shell",
	"sets environment variable named arg0 to arg1",
	"deletes environment variable named arg0",
	"reads file and executes it's contents",
};

int (*builtin_func[]) (char **) = {
	&ush_cd,
	&ush_help,
	&ush_exit,
	&ush_setenv,
	&ush_unsetenv,
	&ush_source,
};

int ush_num_builtins() {
	return sizeof(builtin_str) / sizeof(char *);
}

int ush_cd(char **command) {
	if (command[1] == NULL) {
		fprintf(stderr, "ush: expected argument to \"cd\"\n");
	} else {
		if (chdir(command[1]) != 0) {
			perror("ush");
		}
		ush_update_prompt();
	}
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

int ush_setenv(char **command) {
	if (command[1] == NULL || command[2] == NULL) {
		fprintf(stderr, "ush: expected argument name and value to \"setenv\"\n");
	} else {
		if (setenv(command[1], command[2], 1) != 0) {
			fprintf(stderr, "ush: setenv error\n");
		}
	}
	return 1;
}

int ush_unsetenv(char **command) {
	if (command[1] == NULL) {
		fprintf(stderr, "ush: expected argument to \"unsetenv\"\n");
	} else {
		if (unsetenv(command[1]) != 0) {
			fprintf(stderr, "ush: unsetenv error\n");
		}
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
