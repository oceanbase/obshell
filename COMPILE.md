# Compile

## Install Go environment
[Download and install](https://golang.org/doc/install)

## Supported OS
|OS|Version|Arch|Compile|Execute|
|:---:|:---:|:---:|:---:|:---:|
|CentOS|7.9 / 8.4|x86_64|Yes|Yes|
|Debian|9.13 / 10.11 / 11.2|x86_64|Yes|Yes|
|Fedora|33|x86_64|Yes|Yes|
|OpenSUSE|15.2|x86_64|Yes|Yes|
|Anolis OS|8.2|x86_64|Yes|Yes|
|Ubuntu|16.04 / 18.04 / 20.04|x86_64|Yes|Yes|

## Get the code, build and run
### Clone
Clone the source code to your development machine:
```shell
git clone https://github.com/oceanbase/obshell.git
```
### pre-build

install bindata and swagger:
```shell
go install github.com/go-bindata/go-bindata/...@latest
go install github.com/swaggo/swag/cmd/swag@latest
```
bindata and swagger will be installed in $GOPATH. 

if the $PATH doesn't include %GOPATH in you environment, you need to move the bindata and swagger to you $PATH, or make $PATH include $GOPATH.

Then:
```shell
make pre-build
```

### Build
Build OBShell from the source code in debug mode or release mode:
#### Debug mode
```shell
make build
```
#### Release mode
```shell
make build-release
```