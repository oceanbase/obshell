%define _binaries_in_noarch_packages_terminate_build   0
%define _unpackaged_files_terminate_build              0

# Turn off the brp-python-bytecompile script
%global __os_install_post %(echo '%{__os_install_post}' | sed -e 's!/usr/lib[^[:space:]]*/brp-python-bytecompile[[:space:]].*$!!g')

Name: %(echo ${NAME:oceanbase-ce})
Summary: obshell program
Group: alipay/oceanbase
Version: %(echo $VERSION)
Release: %(echo $RELEASE)%{?dist}
URL: https://github.com/oceanbase/obshell
License: Apache 2.0
BuildArch: x86_64 aarch64 ppc64le
BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)
Prefix: /home/admin

%description
obshell program

%define _prefix /home/admin

%build

%install
RPM_DIR=$OLDPWD
SRC_DIR=$OLDPWD/..
cd $RPM_BUILD_ROOT

if [ %{name} != "obshell" ]; then
    if [ -z "$OB_VERSION" ]; then 
        OB_VERSION=4.2.1.0-100000102023092807
    fi
    OB_version=${OB_VERSION}
    release=$(echo %{?dist})
    [[ $release =~ ^\.([a-zA-Z]+)([0-9]+)$ ]]
    numbers=${BASH_REMATCH[2]}
    rpm2cpio https://mirrors.oceanbase.com/community/stable/el/${numbers}/%{_arch}/oceanbase-ce-${OB_version}.el${numbers}.%{_arch}.rpm| cpio -div
    rpm2cpio https://mirrors.oceanbase.com/community/stable/el/${numbers}/%{_arch}/oceanbase-ce-libs-${OB_version}.el${numbers}.%{_arch}.rpm | cpio -div
    find "./home/admin/oceanbase/bin/" -type f -name "*.py" -exec sed -i '1s_^#!/usr/bin/python$_#!/usr/bin/python2_' {} +
fi

export GOROOT=`go env GOROOT`
export GOPATH=`go env GOPATH`
export PATH=$PATH:$GOROOT/bin
export PATH=$PATH:$GOPATH/bin

cd $SRC_DIR

if [ -n "$OBSHELL_RELEASE" ]; then
    RELEASE=$OBSHELL_RELEASE
fi 

flag="-e VERSION=$VERSION -e RELEASE=$RELEASE -e DIST=%{?dist}"
if [ "$PROXY" ]; then
    flag="$flag -e PROXY=$PROXY"
fi

make pre-build build-release $flag
mkdir -p $RPM_BUILD_ROOT/%{_prefix}/oceanbase/bin
cp bin/obshell $RPM_BUILD_ROOT/%{_prefix}/oceanbase/bin

%files
%defattr(755,admin,admin)
%{_prefix}/oceanbase/*
