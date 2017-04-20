TARGET = ush
SOURCES = $(wildcard *.go)
.PHONY: default all clean
.PRECIOUS: $(TARGET)

default: $(TARGET)
all: default

$(TARGET): $(SOURCES)
	go build -o $(TARGET) $(SOURCES)

install: $(TARGET)
	install ./ush /usr/local/bin/ush

user-install:
	install ./ush ${HOME}/bin/ush

clean:
	-rm -f ush
