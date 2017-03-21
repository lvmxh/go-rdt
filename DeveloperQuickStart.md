# Setup Develoment Environment

## Clone openstackcore-rdtagent code

Make sure you have your $GOPATH, $PATH setup correctly

```
$ cat ~/.bash_rc | grep GOP
declare -x GOPATH=$HOME/go
declare -x PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

$ cd $GOPATH/src
$ git clone https://${your_name}@git-ccr-1.devtools.intel.com/gerrit/p/openstackcore-rdtagent.git
```

## Build & install openstackcore-rdtagent

```

$ go get github.com/tools/godep
$ cd $GOPATH/src/openstackcore-rdtagent
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
