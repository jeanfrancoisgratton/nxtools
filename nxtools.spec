%ifarch aarch64
%global _arch aarch64
%global BuildArchitectures aarch64
%endif

%ifarch x86_64
%global _arch x86_64
%global BuildArchitectures x86_64
%endif

%define debug_package   %{nil}
%define _build_id_links none
%define _name nxtools
%define _prefix /opt
%define _version 0.80.00
%define _rel 0
#%define _arch x86_64
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
#Requires: sudo
#Obsoletes: vmman1 > 1.140

%description
Nexus Repository Manager tools

%prep
%autosetup

%build
cd %{_sourcedir}/%{_name}-%{_version}/src
PATH=$PATH:/opt/go/bin CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -buildid=" -o %{_sourcedir}/%{_binaryname} .

%clean
rm -rf $RPM_BUILD_ROOT

%pre
%install
install -Dpm 0755 %{_sourcedir}/%{_binaryname} %{buildroot}%{_bindir}/%{_binaryname}

%post

%preun

%postun

%files
%defattr(0755,root,root,-)
%{_bindir}/%{_binaryname}


%changelog
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

