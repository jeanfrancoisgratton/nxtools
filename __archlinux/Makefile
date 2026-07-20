# nxtools — Arch Linux (pacman) packaging
#
# Recipes:
#   make build     -> makepkg (build the package into PKGDEST)
#   make release   -> upload the freshly built package to the Nexus archLocal repo (with reindex)
#   make clean     -> remove the makepkg working directories
#   make info      -> print the values extracted from PKGBUILD
#
# Package metadata (name / version / pkgrel / arch / binary) is read straight
# from PKGBUILD so there is a single source of truth.

SHELL := /bin/bash

PKGBUILD  := PKGBUILD

NAME      := $(shell . ./$(PKGBUILD); echo $$pkgname)
VERSION   := $(shell . ./$(PKGBUILD); echo $$pkgver)
PKGREL    := $(shell . ./$(PKGBUILD); echo $$pkgrel)
ARCH      := $(shell . ./$(PKGBUILD); echo $${arch[0]})
BINARY    := $(shell . ./$(PKGBUILD); echo $$_binaryname)

NEXUSREPO := archLocal
PKGDEST   := /data
PKGEXT    := pkg.tar.zst

# makepkg drops the package at: $PKGDEST/<name>-<version>-<pkgrel>-<arch>.<PKGEXT>
PKG       := $(PKGDEST)/$(NAME)-$(VERSION)-$(PKGREL)-$(ARCH).$(PKGEXT)

.PHONY: build release clean info

build:
	rm -rf src pkg
	PKGDEST=$(PKGDEST) makepkg --force --syncdeps --noconfirm

release: build
	/opt/bin/nxtools upload -r $(NEXUSREPO) $(PKG)

clean:
	rm -rf src pkg

info:
	@echo "NAME    = $(NAME)"
	@echo "VERSION = $(VERSION)"
	@echo "PKGREL  = $(PKGREL)"
	@echo "ARCH    = $(ARCH)"
	@echo "BINARY  = $(BINARY)"
	@echo "PKG     = $(PKG)"
