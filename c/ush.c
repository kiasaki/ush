#include "ush.h"

typedef struct Alias Alias;
struct Alias {
	char *name;
	char *value;
	Alias *next;
};

Alias *aliases = NULL;

void ush_add_alias(char *name, char *value) {
	Alias *curr = aliases;
	while (curr->next != NULL) {
		curr = curr->next;
	}

	Alias *new_alias = malloc(sizeof(Alias));
	new_alias->next = NULL;

	new_alias->name = malloc(sizeof(char) * (strlen(name) + 1));
	strcpy(new_alias->name, name);
	new_alias->name[strlen(name)] = '\0';

	new_alias->value = malloc(sizeof(char) * (strlen(value) + 1));
	strcpy(new_alias->value, value);
	new_alias->value[strlen(value)] = '\0';

	curr->next = new_alias;
}

char * volatile ush_prompt;

char *ush_config_file_path(char *filename) {
	char *home_folder = getenv("HOME");
	char *res;
	asprintf(&res, "%s/%s", home_folder, filename);
	return res;
}

void ush_update_cwd(void) {
	char *cwd;
	char buff[PATH_MAX];
	memset(&buff[0], 0, sizeof(char) * PATH_MAX);
	
	cwd = getcwd(buff, PATH_MAX);

	if (cwd == NULL) {
		fprintf(stderr, "ush: getcwd error\n");
		exit(EXIT_FAILURE);
	}

	setenv("PWD", cwd, 1);

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
	ush_update_cwd();

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
			char *error_message = malloc((strlen(command[0])+6) * sizeof(char));
			strcpy(&error_message[0], "ush[");
			strcpy(&error_message[4], command[0]);
			strcpy(&error_message[4+strlen(command[0])], "]\0");
			perror(error_message);
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

	Alias *curr = aliases;
	do {
		if (strcmp(curr->name, command[0]) == 0) {
			char **alias_cmd = ush_parse(curr->value);
			int cmd_len;
			for (cmd_len = 0; command[cmd_len] != NULL; cmd_len++) {}
			int alias_cmd_len;
			for (alias_cmd_len = 0; alias_cmd[alias_cmd_len] != NULL; alias_cmd_len++) {}

			// n elements in alias, n elements in command, minus command[0] which we replaced,
			// plus one for terminating NULL
			char **full_alias_cmd = malloc(sizeof(char *) * (alias_cmd_len + (cmd_len - 1) + 1));
			for (int i = 0; i < alias_cmd_len; i++) {
				full_alias_cmd[i] = alias_cmd[i];
			}
			// skip 1st element
			for (int i = 1; i < cmd_len; i++) {
				full_alias_cmd[alias_cmd_len+i] = command[i];
			}
			full_alias_cmd[alias_cmd_len + (cmd_len - 1)] = NULL;

			return ush_launch(full_alias_cmd);
		}
		curr = curr->next;
	} while (curr != NULL);

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

void ush_run_file(char *filename) {
	long flen;
	char *fcontents;
	FILE *f = fopen(filename, "r");
	if (f == NULL) {
		fprintf(stderr, "ush: can not read file %s\n", filename);
		exit(EXIT_FAILURE);
	}
	
	// go to the end, record length, go back to the start
	fseek(f, 0L, SEEK_END);
	flen = ftell(f);
	fseek(f, 0L, SEEK_SET);

	fcontents = (char*)calloc(flen, sizeof(char));
	if (fcontents == NULL) {
		fprintf(stderr, "ush: allocation error\n");
		exit(EXIT_FAILURE);
	}

	fread(fcontents, sizeof(char), flen, f);
	fclose(f);

	char *line = strtok(fcontents, "\r\n");
	while(line != NULL) {
		char **command = ush_parse(line);
		int result = ush_execute(command);

		free(command);

		if (result == 0) {
			exit(EXIT_SUCCESS);
		}

		line = strtok(NULL, "\r\n");
	}

	free(fcontents);
}

void ush_run_user_config(char *config_filename) {
	char *expanded_filename = ush_config_file_path(config_filename);

	// check the file exists
	FILE *f = fopen(expanded_filename, "r");
	if (f == NULL) {
		return;
	}
	fclose(f);

	ush_run_file(expanded_filename);
	free(expanded_filename);
}

int main(int argc, char **argv) {
    char *program_name = argv[0];

	// Initialize alias list
	aliases = malloc(sizeof(Alias));
	aliases->name = "__ush_alias_sentinel__";
	aliases->value = "__ush_alias_sentinel__";
	aliases->next = NULL;

	bool ran_file = false;

	ush_run_user_config(".ushrc");

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
		} else {
			ran_file = true;

			ush_run_file(*argv);
		}
	}

	if (ran_file) {
		// don't start a session, we just wanted to run a file
		exit(EXIT_SUCCESS);
	}

	ush_run_user_config(".ush_profile");

	ush_init();

	ush_loop();

	return EXIT_SUCCESS;
}
