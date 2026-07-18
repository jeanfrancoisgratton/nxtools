%define debug_package   %{nil}
%define _build_id_links none
%define _name nxtools
%define _prefix /opt
%define _bash_completionsdir /usr/share/bash-completion/completions
%define _zsh_completionsdir  /usr/share/zsh/site-functions
%define _version 1.0.0
%define _rel 1
%define _binaryname nxtools

Name:       nxtools
Version:    %{_version}
Release:    %{_rel}
Summary:    Nexus Repository Manager tools

Group:      CI/CD
License:    GPL2.0
URL:        https://git.famillegratton.net:3000/devops/nxtools.git

Source0:    %{name}-%{_version}.tar.gz
#BuildArchitectures: x86_64
BuildRequires: gcc
Recommends: zsh
Requires: bash-completion

%description
Nexus Repository Manager tools

%prep
%autosetup

%build
cd src
go mod download
PATH=$PATH:/opt/go/bin CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -buildid=" -o %{_builddir}/%{_binaryname} .

%clean
rm -rf $RPM_BUILD_ROOT

%pre

%install
install -Dpm 0755 %{_builddir}/%{_binaryname} %{buildroot}%{_bindir}/%{_binaryname}

%post
# Bash completion — always install
/opt/bin/nxtools completion bash > %{_bash_completionsdir}/nxtools

# Zsh completion — only if zsh is present
if command -v zsh > /dev/null 2>&1; then
    mkdir -p /usr/share/zsh/site-functions
    /opt/bin/nxtools completion zsh > /usr/share/zsh/site-functions/_nxtools
    zsh -c 'autoload -Uz compinit && compinit' 2>/dev/null || true
fi

%preun

%postun
if [ $1 -eq 0 ]; then
    # $1 == 0 means this is a full uninstall, not an upgrade
    rm -f %{_bash_completionsdir}/nxtools
    rm -f %{_zsh_completionsdir}/_nxtools
fi

%files
%defattr(0755,root,root,-)
%{_bindir}/%{_binaryname}


%changelog
* Sat Jul 18 2026 Binary package builder <builder@famillegratton.net> 1.0.0-1
- more rpmbuild enhancements
- decoupled the version subcommand from COBRA
- updated docs, version bump
- test: add unit tests across packages; fix yum deploy-policy check
- reindex: skip Alpine (no rebuild-metadata task; APKINDEX is auto-generated) Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>
- Claude Code integration -- Alpine would not be there without it
- added Alpine repository format
- more apk fixes
- Merge remote-tracking branch 'refs/remotes/origin/develop' into develop
- buildscripts fixes
- Go version bump
- Fixed alpinebuild
- doc update
- arch linux dependencies fixes
- fixed post-inst for zsh
- chore: update changelog for 0.91.00~DEBUG-0
- fixed typo in post-inst
- fixed typo in archlinux post-inst
- Fixed missing wiring of completionCmd to rootCmd
- Release bump
- fixed path in autocompletion tasks
- bumped release number
- chore: update changelog for 0.90.00~DEBUG-1
- Merge branch 'repositories' into develop
- version bump
- Merge branch 'repositories' into develop
- cosmetic changes
- Fixed version number for APK
- release number bump
- chore: update changelog for 0.90.00~DEBUG-0
- fixed yum upload silently disrepecting specs
- Merge branch 'repositories' into develop
- updated build scripts
- Merge branch 'develop' of ssh://git.famillegratton.net:9722/devops/nxtools into develop
- moved a build script in a proper directory
- version bump
- command completion support enabled at install time
- update the list of supported repo formats
- aligned nxtools with RH packaging
- interim submit
- changed perms on file
- Added a build script
- Merge remote-tracking branch 'refs/remotes/origin/develop' into develop
- Merge branch 'repositories' into develop
- interim commit before it gets too messy to merge branches
- Removed un-needed vars from specfile
- prepped nxtools for the new rpmbuilder
- Added ArchLinux packaging support
- Merge pull request 'repositories' (#1) from repositories into develop
- added an helper function, repo supported
- ensure that uploads to raw repos properly handle directories
- Automatic commit of package [nxtools] release [0.80.00-1].
- Minor tweak on showing version number
- Added shorthand flag for directory
- Automatic commit of package [nxtools] release [0.80.00-0].
- Merge branch 'repositories' into develop
- Repos can now be reindexed after uploads
- Automatic commit of package [nxtools] release [0.75.01-0].
- Fixed flag duplicated usage in repo create
- Automatic commit of package [nxtools] release [0.75.00-0].
- Completed all repo formats except docker
- version bump, completed all generic formats
- verbosity fixes, response handling fixed in blob create
- Automatic commit of package [nxtools] release [0.70.00-0].
- Merge branch 'repositories' into develop
- Completed the apt format repo creation
- Version bump
- Automatic commit of package [nxtools] release [0.62.01-0].
- fixed a broken link
- Completed the assets doc
- updated doc
- Fixed assets rm
- links fix so they now (...should) work properly
- doc phase 2
- added a comments field in the environment file, and removed the NEXUS_HOST env var
- Merge branch 'documentation' into develop
- bugfix to asset ls
- Automatic commit of package [nxtools] release [0.62.00-0].
- completed doc, for now
- Merge branch 'develop' into documentation
- added verbosity to repo rm
- foolproofed blob rm
- doc update, version bump
- Merge branch 'develop' into documentation
- Merge branch 'develop' into documentation
- Merge branch 'develop' into blobStores
- fixed path issue in blob ls
- fixed path issue in blob ls
- Merge branch 'develop' into documentation
- Merge branch 'repositories' into develop
- Added the assets count column to repo ls
- interim update
- Merge branch 'develop' into documentation
- Merge branch 'repositories' into develop
- simplified output of repo ls
- interim commit
- Merge branch 'blobStores' into repositories
- preparing to refactor repositories
- doc update
- Automatic commit of package [nxtools] release [0.60.00-0].
- Implement extra info and --json flag for repo ls
- added repositorySettings json payloads doc
- Added extra information to blobstore ls
- Automatic commit of package [nxtools] release [0.50.00-0].
- Version bump, ready to merge
- completed the assets commands
- completed most of assets
- fixed error handling in c.Do()
- Merge branch 'assets' into develop
- Completed the assets commands
- Automatic commit of package [nxtools] release [0.40.00-0].
- version bump
- Major refactoring, completed the assets subcommand
- Added the 'assets ls' subcommand
- types corrections
- Automatic commit of package [nxtools] release [0.30.00-1].
- Merge branch 'repositories' into develop
- Bumped release number for all packages
- Fixed perm on build script
- Automatic commit of package [nxtools] release [0.30.00-0].
- Fixed reindex task
- version bump
- completed repo reindex tasks
- Completed upload
- completed but untested the upload command
- simplified error output
- Automatic commit of package [nxtools] release [0.20.00-0].
- Initialized to use tito.
- Fixed result output on blob create
- Merge branch 'blobStores' into develop
- Completed blobstores
- Merge branch 'main' into develop
- added completion and completed blob rm
- doc update
- added icon image
- go version update
- Implemented blob ls
- refactored the rest client
- minor UX changes
- Completed (not tested) the repositories subpackage
- interim commit
- cosmetic change to headers of all files
- Adapted the http client from dtools to nxtools
- another image fix
- image update
- updated README to properly render logo
- added logos and icons
- adapted code from dtools to nxtools
- builddeps updates
- removed duped files from botched rebase, builddeps update
- Fixed sync to wrong target
- interim commit
- fixed issue where cmd/ had been moved into another package

* Sat May 30 2026 Binary package builder <builder@famillegratton.net> 0.91.00~DEBUG-0
- fixed typo in post-inst
- fixed typo in archlinux post-inst
- Fixed missing wiring of completionCmd to rootCmd
- Release bump
- fixed path in autocompletion tasks
- bumped release number
- chore: update changelog for 0.90.00~DEBUG-1
- Merge branch 'repositories' into develop
- version bump
- Merge branch 'repositories' into develop
- cosmetic changes
- Fixed version number for APK
- release number bump
- chore: update changelog for 0.90.00~DEBUG-0
- fixed yum upload silently disrepecting specs
- Merge branch 'repositories' into develop
- updated build scripts
- Merge branch 'develop' of ssh://git.famillegratton.net:9722/devops/nxtools into develop
- moved a build script in a proper directory
- version bump
- command completion support enabled at install time
- update the list of supported repo formats
- aligned nxtools with RH packaging
- interim submit
- changed perms on file
- Added a build script
- Merge remote-tracking branch 'refs/remotes/origin/develop' into develop
- Merge branch 'repositories' into develop
- interim commit before it gets too messy to merge branches
- Removed un-needed vars from specfile
- prepped nxtools for the new rpmbuilder
- Added ArchLinux packaging support
- Merge pull request 'repositories' (#1) from repositories into develop
- added an helper function, repo supported
- ensure that uploads to raw repos properly handle directories

* Thu May 28 2026 Binary package builder <builder@famillegratton.net> 0.90.00~DEBUG-1
- Merge branch 'repositories' into develop
- version bump
- Merge branch 'repositories' into develop
- cosmetic changes
- Fixed version number for APK
- release number bump
- chore: update changelog for 0.90.00~DEBUG-0
- fixed yum upload silently disrepecting specs
- Merge branch 'repositories' into develop
- updated build scripts
- Merge branch 'develop' of ssh://git.famillegratton.net:9722/devops/nxtools into develop
- moved a build script in a proper directory
- version bump
- command completion support enabled at install time
- update the list of supported repo formats
- aligned nxtools with RH packaging
- interim submit
- changed perms on file
- Added a build script
- Merge remote-tracking branch 'refs/remotes/origin/develop' into develop
- Merge branch 'repositories' into develop
- interim commit before it gets too messy to merge branches
- Removed un-needed vars from specfile
- prepped nxtools for the new rpmbuilder
- Added ArchLinux packaging support
- Merge pull request 'repositories' (#1) from repositories into develop
- added an helper function, repo supported
- ensure that uploads to raw repos properly handle directories

* Thu May 28 2026 Binary package builder <builder@famillegratton.net> 0.90.00~DEBUG-0
- fixed yum upload silently disrepecting specs
- Merge branch 'repositories' into develop
- updated build scripts
- Merge branch 'develop' of ssh://git.famillegratton.net:9722/devops/nxtools into develop
- moved a build script in a proper directory
- version bump
- command completion support enabled at install time
- update the list of supported repo formats
- aligned nxtools with RH packaging
- interim submit
- changed perms on file
- Added a build script
- Merge remote-tracking branch 'refs/remotes/origin/develop' into develop
- Merge branch 'repositories' into develop
- interim commit before it gets too messy to merge branches
- Removed un-needed vars from specfile
- prepped nxtools for the new rpmbuilder
- Added ArchLinux packaging support
- Merge pull request 'repositories' (#1) from repositories into develop
- added an helper function, repo supported
- ensure that uploads to raw repos properly handle directories

* Tue May 12 2026 Binary package builder <builder@famillegratton.net> 0.85.00-0
- interim submit
- changed perms on file
- Added a build script
- Merge remote-tracking branch 'refs/remotes/origin/develop' into develop
- Merge branch 'repositories' into develop
- interim commit before it gets too messy to merge branches
- Removed un-needed vars from specfile
- prepped nxtools for the new rpmbuilder
- Added ArchLinux packaging support
- Merge pull request 'repositories' (#1) from repositories into develop
- added an helper function, repo supported
- ensure that uploads to raw repos properly handle directories

* Wed Mar 25 2026 Binary package builder <builder@famillegratton.net> 0.80.00-1
- Minor tweak on showing version number (jean-francois@famillegratton.net)
- Added shorthand flag for directory (jean-francois@famillegratton.net)

* Wed Mar 25 2026 Binary package builder <builder@famillegratton.net> 0.80.00-0
- Repos can now be reindexed after uploads (jean-francois@famillegratton.net)

* Wed Mar 25 2026 Binary package builder <builder@famillegratton.net> 0.75.01-0
- Fixed flag duplicated usage in repo create (jean-francois@famillegratton.net)

* Wed Mar 25 2026 Binary package builder <builder@famillegratton.net> 0.75.00-0
- Completed all repo formats except docker (jean-francois@famillegratton.net)
- version bump, completed all generic formats (jean-
  francois@famillegratton.net)
- verbosity fixes, response handling fixed in blob create (jean-
  francois@famillegratton.net)

* Tue Mar 24 2026 Binary package builder <builder@famillegratton.net> 0.70.00-0
- Completed the apt format repo creation (jean-francois@famillegratton.net)
- Version bump (jean-francois@famillegratton.net)

* Mon Mar 23 2026 Binary package builder <builder@famillegratton.net> 0.62.01-0
- fixed a broken link (jean-francois@famillegratton.net)
- Completed the assets doc (jean-francois@famillegratton.net)
- updated doc (jean-francois@famillegratton.net)
- Fixed assets rm (jean-francois@famillegratton.net)
- links fix so they now (...should) work properly (jean-
  francois@famillegratton.net)
- doc phase 2 (jean-francois@famillegratton.net)
- added a comments field in the environment file, and removed the NEXUS_HOST
  env var (jean-francois@famillegratton.net)
- bugfix to asset ls (jean-francois@famillegratton.net)

* Sat Mar 21 2026 Binary package builder <builder@famillegratton.net> 0.62.00-0
- completed doc, for now (jean-francois@famillegratton.net)
- added verbosity to repo rm (jean-francois@famillegratton.net)
- foolproofed blob rm (jean-francois@famillegratton.net)
- doc update, version bump (jean-francois@famillegratton.net)
- fixed path issue in blob ls (jean-francois@famillegratton.net)
- fixed path issue in blob ls (jean-francois@famillegratton.net)
- Added the assets count column to repo ls (jean-francois@famillegratton.net)
- interim update (jean-francois@famillegratton.net)
- simplified output of repo ls (jean-francois@famillegratton.net)
- interim commit (jean-francois@famillegratton.net)
- preparing to refactor repositories (jean-francois@famillegratton.net)
- doc update (jean-francois@famillegratton.net)

* Thu Mar 19 2026 Binary package builder <builder@famillegratton.net> 0.60.00-0
- Implement extra info and --json flag for repo ls (jean-
  francois@famillegratton.net)
- added repositorySettings json payloads doc (jean-francois@famillegratton.net)
- Added extra information to blobstore ls (jean-francois@famillegratton.net)

* Mon Mar 16 2026 Binary package builder <builder@famillegratton.net> 0.50.00-0
- Version bump, ready to merge (jean-francois@famillegratton.net)
- completed the assets commands (jean-francois@famillegratton.net)
- completed most of assets (jean-francois@famillegratton.net)
- fixed error handling in c.Do() (jean-francois@famillegratton.net)
- Completed the assets commands (jean-francois@famillegratton.net)

* Fri Mar 13 2026 Binary package builder <builder@famillegratton.net> 0.40.00-0
- version bump (jean-francois@famillegratton.net)
- Major refactoring, completed the assets subcommand (jean-
  francois@famillegratton.net)
- Added the 'assets ls' subcommand (jean-francois@famillegratton.net)
- types corrections (jean-francois@famillegratton.net)

* Tue Mar 10 2026 Binary package builder <builder@famillegratton.net> 0.30.00-1
- Bumped release number for all packages (jean-francois@famillegratton.net)
- Fixed perm on build script (builder@famillegratton.net)

* Tue Mar 10 2026 Binary package builder <builder@famillegratton.net> 0.30.00-0
- Fixed reindex task (jean-francois@famillegratton.net)
- version bump (jean-francois@famillegratton.net)
- completed repo reindex tasks (jean-francois@famillegratton.net)
- Completed upload (jean-francois@famillegratton.net)
- completed but untested the upload command (jean-francois@famillegratton.net)
- simplified error output (jean-francois@famillegratton.net)

* Sun Mar 08 2026 Binary package builder <builder@famillegratton.net> 0.20.00-0
- new package built with tito

