PROJECT_ROOT := src/
VERSION = 0.1

.DEFAULT_GOAL := all

# The resulting binaries map to the subproject names
BINARIES = \
	solbuild

include Makefile.gobuild

_PKGS = \
	solbuild \
	solbuild/cmd

# We want to add compliance for all built binaries
_CHECK_COMPLIANCE = $(addsuffix .compliant,$(_PKGS))

# Build all binaries as dynamic binary
BINS = $(addsuffix .dynbin,$(BINARIES))

# Ensure our own code is compliant..
compliant: $(_CHECK_COMPLIANCE)
install: $(BINS)
	test -d $(DESTDIR)/usr/bin || install -D -d -m 00755 $(DESTDIR)/usr/bin; \
	install -m 00755 builds/* $(DESTDIR)/usr/bin/.

ensure_modules:
	@ ( \
		git submodule init; \
		git submodule update; \
	);

release:
	git archive --format=tar.gz --verbose -o solbuild-$(VERSION).tar.gz HEAD --prefix=solbuild-$(VERSION)/

all: $(BINS)
