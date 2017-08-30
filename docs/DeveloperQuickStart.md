# Setup Develoment Environment

## Prepare GO development environment

Make sure you have your $GOPATH, $PATH setup correctly

```
e.g.
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
```

## Clone openstackcore-rdtagent code

Clone the code into $GOPATH/src/openstackcore-rdtagent

## Build & install openstackcore-rdtagent

```
$ go get github.com/tools/godep

(goto source code topdir)
$ git submodule init
$ git submodule update
$ ./install-deps
```

## Run openstackcore-rdtagent

```
$ $GOPATH/bin/openstackcore-rdtagent --help
$ $GOPATH/bin/openstackcore-rdtagent
```

## Commit code

There's a bash shell script `hacking.sh` will help to checking coding style
by `go fmt` and `golint`.

After you commit your changes, run `./hacking.sh` and address errors before
push your changes.

## Test

There's a bash shell script `test.sh`, a helper scirpt to do uint testing and
functional testing.

`./test.sh -u` to run all unit test cases.
`./test.sh -i` to run all functional test cases.

To read test.sh to understand what functional test case do.

## Godep

Check [ Godep ](https://github.com/tools/godep) for how to add/update dependencies.

## Swagger

The API defination located under docs/v1/swagger.yaml

Upload docs/api/v1/swagger.yaml to http://editor.swagger.io/#!/ , it will help to generate client.
