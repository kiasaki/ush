#include "ush.h"

char **ush_parse(char *line) {
	int bufsize = USH_TOKEN_BUFER_SIZE;
	int position = 0;
	char **tokens = malloc(bufsize * sizeof(char*));
	char *token;

	if (!tokens) {
		fprintf(stderr, "ush: allocation error\n");
		exit(EXIT_FAILURE);
	}

	token = strtok(line, USH_TOKEN_DELIMITER);
	while (token != NULL) {
		tokens[position] = token;
		position++;

		// grow buffersize to fit token
		if (position >= bufsize) {
			bufsize += USH_TOKEN_BUFER_SIZE;
			tokens = realloc(tokens, bufsize * sizeof(char*));
			if (!tokens) {
				fprintf(stderr, "lsh: allocation error\n");
				exit(EXIT_FAILURE);
			}
		}

		token = strtok(NULL, USH_TOKEN_DELIMITER);
	}

	tokens[position] = NULL;
	return tokens;
}
