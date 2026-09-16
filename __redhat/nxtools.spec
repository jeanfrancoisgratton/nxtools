%define debug_package   %{nil}
%define _build_id_links none
%define _name nxtools
%define _prefix /opt
%define _version 1.3.3
%define _rel 3
%define _arch x86_64
%define _binaryname nxtools

Name:       nxtools
Version:    %{_version}
Release:    %{_rel}
Summary:    Nexus Repository Management tools

Group:      Packaging tool
License:    GPL2.0
URL:        https://git.famillegratton.net:3000/devops/nxtools

Source0:    %{name}-%{_version}.tar.gz
#BuildArchitectures: x86_64
BuildRequires: gcc
#Requires: sudo
#Obsoletes: vmman1 > 1.140

%description
Nexus Repository Management tools

%prep
%autosetup

%build
cd src
go mod download
# rpmbuild runs %build under set -e, so both of these fast-fail the package
# build; go test exits 0 for packages with no test files and only fails on
# an actual test failure.
PATH=$PATH:/opt/go/bin CGO_ENABLED=0 go vet ./...
PATH=$PATH:/opt/go/bin CGO_ENABLED=0 go test ./...
PATH=$PATH:/opt/go/bin CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -buildid= -X nxtools/cmd.buildVersion=%{_version} -X nxtools/cmd.buildDate=%(date +%%Y.%%m.%%d)" -o %{_builddir}/%{name}-%{version}/%{_binaryname} .

%clean
rm -rf $RPM_BUILD_ROOT

%pre

%install
rm -rf %{buildroot}
install -Dpm 0755 %{_builddir}/%{name}-%{version}/%{_binaryname} %{buildroot}%{_bindir}/%{_binaryname}

%post

%preun

%postun

%files
%defattr(0755,root,root,-)
%{_bindir}/%{_binaryname}


%changelog
* Wed Sep 16 2026 Binary package builder <builder@famillegratton.net> 1.3.3-2
- chore: version bump

* Wed Sep 16 2026 Binary package builder <builder@famillegratton.net> 1.3.3-1
- version bump
- completing merge
- Merge branch 'task_management' into develop
- feat(TASKS): Completed tasks create blob-compact
- feature: stubbed a new task command
- chore: update changelog for 1.3.2-1

* Tue Sep 15 2026 Binary package builder <builder@famillegratton.net> 1.3.2-1
- chore: package version bump, doc update
- enhancement: apt repo creation/migration now supports automated signing key generation
- bug: alpine migration to a non-existing repo failed because --sign was missing
- arch is now part of the version string
- Merge branch 'main' into develop
- chore: update changelog for 1.3.1-1
- chore: update changelog for 1.3.0-2
- chore: update changelog for 1.3.0-1
- chore: update changelog for 1.2.0-1

* Tue Sep 15 2026 Binary package builder <builder@famillegratton.net> 1.3.1-1
- enhancement: build fails on go vet/go test failure; dynamic version numbering change scheme
- chore: another mislabeled version fix

* Tue Sep 15 2026 Binary package builder <builder@famillegratton.net> 1.3.0-2
- chore: another mislabeled version fix

* Tue Sep 15 2026 Binary package builder <builder@famillegratton.net> 1.3.0-1
- Last release was mis-versioned
- chore: removed dontexec flag
- bug: fixed showstoppers that broke repos
- builddeps update
- Merge remote-tracking branch 'refs/remotes/origin/develop' into develop
- chore: removed shell completion from packaging, GO version bump, builddeps update, prepping for software version bump
- chore: moved repo migrate in its own package for further work
- chore: update changelog for 1.2.1-2
- RPMBUILDER: record the RPM changelog on develop instead of main

* Tue Sep 15 2026 Binary package builder <builder@famillegratton.net> 1.2.0-1
- bug: fixed showstoppers that broke repos
- builddeps update
- Merge remote-tracking branch 'refs/remotes/origin/develop' into develop
- chore: removed shell completion from packaging, GO version bump, builddeps update, prepping for software version bump
- chore: moved repo migrate in its own package for further work
- chore: update changelog for 1.2.1-2
- RPMBUILDER: record the RPM changelog on develop instead of main

* Fri Jul 24 2026 Binary package builder <builder@famillegratton.net> 1.1.4-1
- added assets fetch support for raw-format repos

* Thu Jul 23 2026 Binary package builder <builder@famillegratton.net> 1.1.3-3
- Other makefile fixes
- Merge remote-tracking branch 'refs/remotes/origin/develop' into develop
- fixed tagging issue in RPMBUILD
- chore: update changelog for 1.1.3-2
- fixed multiple post-inst issues
- chore: update changelog for 1.1.3-1
- version bump
- fixed the signing keys generation for Alpine repos
- Revert "Enforce PKCS1 key format to sign APK packages"
- removed un-needed scripts

* Thu Jul 23 2026 Binary package builder <builder@famillegratton.net> 1.1.3-2
- fixed multiple post-inst issues
- chore: update changelog for 1.1.3-1
- version bump
- fixed the signing keys generation for Alpine repos
- Revert "Enforce PKCS1 key format to sign APK packages"
- removed un-needed scripts

* Thu Jul 23 2026 Binary package builder <builder@famillegratton.net> 1.1.3-1
- version bump
- fixed the signing keys generation for Alpine repos
- Revert "Enforce PKCS1 key format to sign APK packages"
- removed un-needed scripts

* Thu Jul 23 2026 Binary package builder <builder@famillegratton.net> 1.1.2-1
- Merge remote-tracking branch 'refs/remotes/origin/develop' into develop
- refreshed version
- Enforce PKCS1 key format to sign APK packages

