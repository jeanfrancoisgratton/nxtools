%define debug_package   %{nil}
%define _build_id_links none
%define _name nxtools
%define _prefix /opt
%define _bash_completionsdir /usr/share/bash-completion/completions
%define _zsh_completionsdir  /usr/share/zsh/site-functions
%define _version 1.1.4
%define _rel 1
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
PATH=$PATH:/opt/go/bin CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -buildid=" -o %{_builddir}/%{name}-%{version}/%{_binaryname} .

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

