PROJECT_ROOT := src/
VERSION = 0.1

.DEFAULT_GOAL := all

# The resulting binaries map to the subproject names
BINARIES = \
	solbuild

GO_TESTS = \
	builder.test

include Makefile.gobuild

_PKGS = \
	builder \
	solbuild \
	solbuild/cmd

# We want to add compliance for all built binaries
_CHECK_COMPLIANCE = $(addsuffix .compliant,$(_PKGS))

# Build all binaries as static binary
BINS = $(addsuffix .statbin,$(BINARIES))

# Ensure our own code is compliant..
compliant: $(_CHECK_COMPLIANCE)
install: $(BINS)
	test -d $(DESTDIR)/usr/bin || install -D -d -m 00755 $(DESTDIR)/usr/bin; \
	install -m 00755 bin/* $(DESTDIR)/usr/bin/.; \
	test -d $(DESTDIR)/usr/share/solbuild || install -D -d -m 00755 $(DESTDIR)/usr/share/solbuild; \
	install -m 00644 data/*.profile $(DESTDIR)/usr/share/solbuild/.;
	install -m 00644 data/solbuild.conf $(DESTDIR)/usr/share/solbuild/.;

ensure_modules:
	@ ( \
		git submodule init; \
		git submodule update; \
	);

# See: https://github.com/meitar/git-archive-all.sh/blob/master/git-archive-all.sh
release: ensure_modules
	git-archive-all.sh --format tar.gz --prefix solbuild-$(VERSION)/ --verbose -t HEAD solbuild-$(VERSION).tar.gz

all: $(BINS)
