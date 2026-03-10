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
%define _version 0.30.00
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
* Sun Mar 08 2026 Binary package builder <builder@famillegratton.net> 0.20.00-0
- new package built with tito

