%define debug_package   %{nil}
%define _build_id_links none
%define _name nxtools
%define _prefix /opt
%define _bash_completionsdir /usr/share/bash-completion/completions
%define _zsh_completionsdir  /usr/share/zsh/site-functions
%define _version 1.1.3
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
# Bash completion — always install
%{_bindir}/%{_binaryname} completion bash > %{_bash_completionsdir}/{_binaryname}

# Zsh completion — only if zsh is present
if command -v zsh > /dev/null 2>&1; then
    mkdir -p %{_zsh_completionsdir}/zsh/site-functions
    %{_bindir}/%{_binaryname} completion zsh > %{_zsh_completionsdir}/_{_binaryname}
fi

%preun

%postun
if [ $1 -eq 0 ]; then
    # $1 == 0 means this is a full uninstall, not an upgrade
    rm -f %{_bash_completionsdir}/{_binaryname
    rm -f %{_zsh_completionsdir}/_{_binaryname
fi

%files
%defattr(0755,root,root,-)
%{_bindir}/%{_binaryname}


%changelog
* Thu Jul 23 2026 Binary package builder <builder@famillegratton.net> 1.1.2-1
- Merge remote-tracking branch 'refs/remotes/origin/develop' into develop
- refreshed version
- Enforce PKCS1 key format to sign APK packages

