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
$ ./install
```

## Run openstackcore-rdtagent

```
$ $GOPATH/bin/openstackcore-rdtagent --help
$ $GOPATH/bin/openstackcore-rdtagent
```


## Testing

In another terminal :

```
$ curl -i localhost:8081/v1/cpuinfo

```


## Godep

Check [ Godep ](https://github.com/tools/godep) for how to add/update dependencies.

## Unit Test

TODO


Happy hacking!

## Swagger

The API defination located under docs/v1/swagger.yaml

Upload docs/api/v1/swagger.yaml to http://editor.swagger.io/#!/ , it will help to generate client.
