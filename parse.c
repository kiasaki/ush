#include "ush.h"

char **ush_parse(char *line) {
	wordexp_t webuff;
	if (wordexp(line, &webuff, 0) != 0) {
		fprintf(stderr, "ush: word expansion error\n");
		exit(EXIT_FAILURE);
	}

	char **command = malloc((webuff.we_wordc + 1) * sizeof(char *));
	for (size_t i = 0; i < webuff.we_wordc; i++) {
		command[i] = malloc((strlen(webuff.we_wordv[i]) + 1) * sizeof(char));
		strcpy(command[i], webuff.we_wordv[i]);
	}
	command[webuff.we_wordc] = NULL;

	wordfree(&webuff);

	return command;
}
