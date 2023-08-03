# Yashan Trace Collector

## Summary

*Yashan Trace Collector* is a lightweight information collection tool, which can improve the efficiency of operation and maintenance information collection and troubleshooting.

## Features

1. Collect Trace Information With One-click
    * Basic Information Collection(Including: Host And YashanDB)
    * Troubleshooting Information Collection
    * Performance Tuning Information Collection
2. Collect Trace Information As Scheduled
3. Collect Strategy And Report Management

## APIs

// TODO:

## Uasge

// TODO:

## For Development

### Dependencies

#### Development And Compilation Tools

| TOOL   | VERSION |
| ------ | ------- |
| go     | go 1.19 |
| python | python3 |
| make   | v3+     |
| gcc    | v7+     |

##### go

```bash
# example for x86_64
cd ~
# x86_64
wget https://dl.google.com/go/go1.19.7.linux-amd64.tar.gz

# aarch64 
# wget https://dl.google.com/go/go1.19.7.linux-arm64.tar.gz

tar -xzvf go1.19.7.linux-amd64.tar.gz
mkdir golang
mv go golang/

vim ~/.golang.env
<insert>
export GOROOT=${HOME}/golang/go
export GOPATH=${HOME}/golang/src
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct
export PATH=$PATH:${GOROOT}/bin:${GOPATH}/bin
<wq>

vim ~/.bashrc
<insert>
source ~/.golang.env
<wq>

source ~/.bashrc
```

##### python

```bash
# example for centos 7
yum install -y python3
```

##### make

```bash
# example for centos 7 x86_64
yum install -y make.x86_64
```

##### gcc

```bash
# example for centos 7 x86_64
yum install -y centos-release-scl
yum install -y devtoolset-7-gcc.x86_64
yum install -y devtoolset-7-gcc-c++.x86_64
```

#### Code Checking Tools(Optional)

| TOOL          | VERSION |
| ------------- | ------- |
| golangci-lint | v1.53.2 |
| yapf          | 0.32.0  |
| mypy          | 0.950   |
| shellcheck    | 0.3.8   |

##### golangci-lint

```bash

# use script for installation
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh |\
sh -s -- -b $(go env GOPATH)/bin v1.53.2

# or use package for installation
# Releases: https://github.com/golangci/golangci-lint/releases/tag/v1.53.2
# Available Resource: wget http://192.168.8.236:8888/golang/golangci-lint-1.53.2-linux-amd64.tar.gz

# NOTE: such go install/go get installation aren't guaranteed to work. We recommend using binary installation.
```

##### yapf

```bash
pip3 install yapf==0.32.0
```

##### mypy

```bash
pip3 install mypy==0.950
```

##### shellcheck

```bash
# example for centos 7
yum install -y epel-release
yum install -y ShellCheck-0.3.8-1.el7
```

### Configurations

#### Private Repository Configuration

#### go env

```bash
go env -w GOPRIVATE=git.yasdb.com
```

#### git config

```bash
vim ~/.gitconfig
[url "git@git.yasdb.com:"]
    insteadOf = https://git.yasdb.com/
```

### Compilation Script

There's a `build.py` file located at the root directory of the project.

```bash
# clean compilation data
python3 build.py clean

# start compilation
python3 build.py build
# --skip-check  build without checking code (default: False)
# --skip-test   build without running unit test (default: False)
# -c, --clean   clean before building (default: False)
# -f, --force   clean before building, then build without checking code and running unit test (default: False)

# check code
python3 build.py check

# run unit test
python3 build.py test
```

### Quick Start

```bash
cd yashan-trace-collector
export YTC_HOME=$(pwd)
export YTC_DEBUG_MODE=true

go run cmd/ytcctl/*.go

go run cmd/ytcd/*.go
```
