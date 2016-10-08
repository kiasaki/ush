TARGET = ush
LIBS = -lm
CC = gcc
CFLAGS = -g -Wall -pedantic -std=c99

.PHONY: default all clean

default: $(TARGET)
all: default

OBJECTS = $(patsubst %.c, %.o, $(wildcard *.c))
HEADERS = $(wildcard *.h)

%.o: %.c $(HEADERS)
	$(CC) $(CFLAGS) -c $< -o $@

.PRECIOUS: $(TARGET) $(OBJECTS)

$(TARGET): $(OBJECTS)
	$(CC) $(OBJECTS) -Wall -o $@

install:
	install ./ush /usr/local/bin/ush

clean:
	-rm -f *.o
	-rm -f $(TARGET)
