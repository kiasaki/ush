#include "ush.h"

char * volatile ush_prompt;

char *ush_config_file_path(char *filename) {
	char *home_folder = getenv("HOME");
	char *res;
	asprintf(&res, "%s/%s", home_folder, filename);
	return res;
}

void ush_update_prompt(void) {
	char* cwd;
	char buff[PATH_MAX + 1];

	cwd = getcwd(buff, PATH_MAX + 1);

	if (cwd == NULL) {
		fprintf(stderr, "ush: getcwd error\n");
		exit(EXIT_FAILURE);
	}

	int cwdlen = strlen(cwd);
	if (cwd[cwdlen] == '/') {
		cwd[cwdlen] = '\0';
	}
	char *cwd_folder_start = strrchr(cwd, '/');
	if (cwd_folder_start != 0 && cwdlen > 1) {
		cwd_folder_start += 1;
	}

	snprintf(ush_prompt, MAX_PROMPT_SZ, "%s$ ", cwd_folder_start);
}

void ush_completion(const char *buf, linenoiseCompletions *lc) {
	char *last_command_argument = strrchr(buf, ' ');
	if (last_command_argument == NULL) {
		// just start from the beginning
		last_command_argument = buf;
	} else if (strlen(last_command_argument) > 0)  {
		// move pointer past prefixed space
		last_command_argument += 1;
	}

	char glob_path[strlen(last_command_argument) + 3];
	sprintf(glob_path, "%s**", last_command_argument);

	wordexp_t webuff;
	if (wordexp(glob_path, &webuff, WRDE_NOCMD) != 0) {
		fprintf(stderr, "ush: word expansion error\n");
		exit(EXIT_FAILURE);
	}

	for (size_t i = 0; i < webuff.we_wordc; i++) {
		if (strcmp(webuff.we_wordv[i], glob_path) == 0) {
			// when the expanded string is the same as input
			continue;
		}

		int command_start_length = strlen(buf) - strlen(last_command_argument);
		int match_length = strlen(webuff.we_wordv[i]);
		char full_completion[command_start_length + match_length + 1];
		for (int j = 0; j < command_start_length; j++) {
			full_completion[j] = buf[j];
		}
		full_completion[command_start_length] = '\0';
		strcat(full_completion, webuff.we_wordv[i]);

		linenoiseAddCompletion(lc, full_completion);
	}

	wordfree(&webuff);
}

void ush_init(void) {
	linenoiseSetMultiLine(1);
	linenoiseHistorySetMaxLen(10000);
	linenoiseSetCompletionCallback(ush_completion);

	ush_prompt = malloc(MAX_PROMPT_SZ);
	ush_update_prompt();

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
	char *line = linenoise(ush_prompt);

    while (true) {
		if (line != NULL) {
			linenoiseHistoryAdd(line);
			linenoiseHistorySave(history_file);

			char **command = ush_parse(line);
			int result = ush_execute(command);
			free(line);
			free(command);
			if (result == 0) {
				goto cleanup;
			}
		}

		line = linenoise(ush_prompt);
	}

cleanup:
	free(history_file);
	free(ush_prompt);
	exit(EXIT_SUCCESS);
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
