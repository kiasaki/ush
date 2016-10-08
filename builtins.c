#include "ush.h"

char *builtin_str[] = {
	"cd",
	"help",
	"exit"
};

char *builtin_help[] = {
	"changes current directory",
	"prints builtin functions",
	"exits the shell"
};

int (*builtin_func[]) (char **) = {
	&ush_cd,
	&ush_help,
	&ush_exit
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
	}
	return 1;
}

int ush_help(char **command) {
	int i;
	printf("ush builtin functions:\n\n");
	for (i = 0; i < ush_num_builtins(); i++) {
		printf("  %s\t%s\n", builtin_str[i], builtin_help[i]);
	}
	return 1;
}

int ush_exit(char **command) {
	return 0;
}
