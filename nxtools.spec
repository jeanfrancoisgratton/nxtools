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
%define _version 0.50.00
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

