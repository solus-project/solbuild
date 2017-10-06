PROJECT_ROOT := src/
VERSION = 1.4.1

.DEFAULT_GOAL := all

# The resulting binaries map to the subproject names
BINARIES = \
	solbuild

GO_TESTS = \
	builder.test

include Makefile.gobuild

_PKGS = \
	builder \
	builder/source \
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
	install -m 00644 data/00_solbuild.conf $(DESTDIR)/usr/share/solbuild/.;
	test -d $(DESTDIR)/usr/share/man/man1 || install -D -d -m 00755 $(DESTDIR)/usr/share/man/man1; \
	install -m 00644 man/*.1 $(DESTDIR)/usr/share/man/man1/.; \
	test -d $(DESTDIR)/usr/share/man/man5 || install -D -d -m 00755 $(DESTDIR)/usr/share/man/man5; \
	install -m 00644 man/*.5 $(DESTDIR)/usr/share/man/man5/.;


ensure_modules:
	@ ( \
		git submodule init; \
		git submodule update; \
	);

# Credit to swupd developers: https://github.com/clearlinux/swupd-client
MANPAGES = \
	man/solbuild.1 \
	man/solbuild.conf.5 \
	man/solbuild.profile.5

gen_docs:
	for MANPAGE in $(MANPAGES); do \
		ronn --roff < $${MANPAGE}.md > $${MANPAGE}; \
		ronn --html < $${MANPAGE}.md > $${MANPAGE}.html; \
	done

# See: https://github.com/meitar/git-archive-all.sh/blob/master/git-archive-all.sh
release: ensure_modules
	git-archive-all.sh --format tar.gz --prefix solbuild-$(VERSION)/ --verbose -t HEAD solbuild-$(VERSION).tar.gz

all: $(BINS)
