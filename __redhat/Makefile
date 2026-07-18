# Run recipes under bash with pipefail so a failing command in a pipeline
# (e.g. `git archive | gzip`) fails the whole line instead of being masked
# by the exit status of the last command in the pipe.
SHELL       := /bin/bash
.SHELLFLAGS := -o pipefail -c

SPEC      := nxtools.spec
NAME      := $(shell rpmspec -q --qf '%{name}\n'    $(SPEC) | head -1)
VERSION   := $(shell rpmspec -q --qf '%{version}\n' $(SPEC) | head -1)
RELEASE   := $(shell rpmspec -q --qf '%{release}\n' $(SPEC) | head -1)
SOURCES   := $(HOME)/rpmbuild/SOURCES
RPMS      := $(HOME)/rpmbuild/RPMS
DEST      := /data
REPOROOT  := $(shell git rev-parse --show-toplevel)
NEXUSREPO := dnfLocal
TAG       := v$(VERSION)
TARBALL   := $(SOURCES)/$(NAME)-$(VERSION).tar.gz

.PHONY: tag release tarball check-tags changelog rpm rpmcl upload commitcl clean FORCE

# If a recipe fails partway through, delete the (possibly truncated) target
# file it was building, so a corrupt artifact isn't left looking up-to-date.
.DELETE_ON_ERROR:

# Intended workflow: make release

# Real file target: this is what .DELETE_ON_ERROR protects. The FORCE
# prerequisite keeps the previous behaviour of regenerating the tarball on
# every invocation (important when rebuilding the same VERSION after amending).
$(TARBALL): FORCE
	git -C $(REPOROOT) archive --prefix=$(NAME)-$(VERSION)/ HEAD \
	  | gzip > $@

# Friendly alias so `make tarball` still works.
tarball: $(TARBALL)

FORCE:

check-tags:
	@# 1. The condition that silently breaks changelog generation:
	@if git describe --exact-match --tags HEAD >/dev/null 2>&1; then \
	  echo "WARNING: HEAD is already tagged $$(git describe --exact-match --tags HEAD)."; \
	  echo "         updateChangelog.sh will fall back to the previous tag for the range."; \
	fi
	@# 2. Local/remote tag agreement:
	@local=$$(git rev-parse -q --verify "refs/tags/$(TAG)" 2>/dev/null); \
	 remote=$$(git ls-remote --tags origin "refs/tags/$(TAG)" 2>/dev/null | cut -f1); \
	 if   [ -n "$$local" ] && [ -z "$$remote" ]; then echo "NOTE: $(TAG) exists locally but is not pushed."; \
	 elif [ -z "$$local" ] && [ -n "$$remote" ]; then echo "NOTE: $(TAG) is on origin but not local (run: git fetch --tags)."; \
	 elif [ -n "$$local" ] && [ "$$local" != "$$remote" ]; then \
	   echo "ERROR: $(TAG) points to different commits locally ($$local) and on origin ($$remote)."; \
	   exit 1; \
	 fi

changelog: check-tags $(TARBALL)
	$(REPOROOT)/__redhat/updateChangelog.sh

rpm: $(TARBALL)
	rpmbuild -bb $(SPEC)
	find $(RPMS) -name '*.rpm' -exec cp {} $(DEST)/ \;

rpmcl: $(TARBALL) changelog
	rpmbuild -bb $(SPEC)
	find $(RPMS) -name '*.rpm' -exec cp {} $(DEST)/ \;

upload:
	nxtools upload -r $(NEXUSREPO) $(DEST)/$(NAME)-$(VERSION)-$(RELEASE).x86_64.rpm

commitcl:
	git add ${SPEC}
	git commit -m "chore: update changelog for $(VERSION)-$(RELEASE)"

clean:
	rm -f  $(TARBALL)
	rm -rf $(HOME)/rpmbuild/BUILD/$(NAME)-$(VERSION)

tag:
	git tag -a v${VERSION} -m 'v${VERSION}'
	git push origin v${VERSION}

release:
	${MAKE} rpmcl
	${MAKE} commitcl
	git push
	${MAKE} upload
	${MAKE} tag
