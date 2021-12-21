TARGET = ush
SOURCES = $(wildcard *.go)
.PHONY: default all clean
.PRECIOUS: $(TARGET)

default: $(TARGET)
all: default

VERSION = devel
ifneq ($(wildcard .git),)
  # get hash of current commit
  GITREV := $(shell git log -n1 --format="%h")
  # get tag for current commit
  GITTAG := $(shell git tag --contains $(GITREV) | tr '[:upper:]' '[:lower:]')
  ifneq ($(GITTAG),)
	VERSION = $(GITTAG)
  else
    # current commit is not tagged
    # -> include current commit hash in version information
    VERSION = $(GITREV)
  endif
endif


$(TARGET): $(SOURCES)
	go build -ldflags "-X main.ushVersion=$(VERSION)" -o $(TARGET) .

install: $(TARGET)
	install ./ush /usr/local/bin/ush

user-install:
	install ./ush ${HOME}/bin/ush

clean:
	-rm -f ush
