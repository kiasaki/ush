#include "ush.h"

char *ush_config_file_path(char *filename) {
	char *home_folder = getenv("HOME");
	char *res;
	asprintf(&res, "%s/%s", home_folder, filename);
	return res;
}

void ush_completion(const char *buf, linenoiseCompletions *lc) {
	if (buf[0] == 'h') {
		linenoiseAddCompletion(lc,"hello");
	}
}

void ush_init(void) {
	linenoiseSetMultiLine(1);
	linenoiseHistorySetMaxLen(10000);
	// linenoiseSetCompletionCallback(ush_completion);

	char *history_file = ush_config_file_path(".ush_history");
	linenoiseHistoryLoad(history_file);
	free(history_file);
}

int ush_launch(char **command) {
	pid_t pid, wpid;
	int status;

	pid = fork();
	if (pid == 0) {
		// Child process
		if (execvp(command[0], command) == -1) {
			perror("ush");
		}
		exit(EXIT_FAILURE);
	} else if (pid < 0) {
		// Error forking
		perror("ush");
	} else {
		// Parent process
		do {
			wpid = waitpid(pid, &status, WUNTRACED);
		} while (!WIFEXITED(status) && !WIFSIGNALED(status));
	}

	return 1;
}

int ush_execute(char **command) {
	int i;
	if (command[0] == NULL) {
		// No command given
		return 1;
	}

	for (i = 0; i < ush_num_builtins(); i++) {
		if (strcmp(command[0], builtin_str[i]) == 0) {
			return (*builtin_func[i])(command);
		}
	}

	// If we didn't match with a builtin, exec command
	return ush_launch(command);
}

void ush_loop(void) {
	char *history_file = ush_config_file_path(".ush_history");
	char *prompt = "$ ";
	char *line;
    while((line = linenoise(prompt)) != NULL) {
		linenoiseHistoryAdd(line);
		linenoiseHistorySave(history_file);

		char **command = ush_parse(line);
		int result = ush_execute(command);
		free(line);
		free(command);
		if (result == 0) {
			free(history_file);
			exit(EXIT_SUCCESS);
		}
	}
	free(history_file);
}

int main(int argc, char **argv) {
    char *program_name = argv[0];

	while(argc > 1) {
		argc--;
		argv++;
		if (!strcmp(*argv,"--version") || !strcmp(*argv,"-v")) {
			fprintf(stderr, "ush %s\n", USH_VERSION);
			exit(EXIT_SUCCESS);
		} else if (!strcmp(*argv,"--help") || !strcmp(*argv,"-h")) {
			fprintf(stderr, "Usage: %s [<argument> ...]\n\n", program_name);
			fprintf(stderr, "Special options:\n");
			fprintf(stderr, "  --help    show this message, then exit\n");
			fprintf(stderr, "  --version show ush version number, then exit\n");
			exit(EXIT_SUCCESS);
		}
	}

	ush_init();

	ush_loop();

	return EXIT_SUCCESS;
}
