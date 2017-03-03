# Setup Develoment Environment

## Clone openstackcore-rdtagent code

Make sure you have your $GOPATH setup correctly

``
git clone https://${your_name}@git-ccr-1.devtools.intel.com/gerrit/p/openstackcore-rdtagent.git $GOPATH/src
```

## Build & install openstackcore-rdtagent

```
cd $GOPATH/src/openstackcore-rdtagent

./install

```

## Run openstackcore-rdtagent

```
$GOPATH/bin/openstackcore-rdtagent
```


## Testing

In another terminal :

```
curl -i localhost:8081/v1/cpuinfo

```


## Godep

Check [ Godep ](https://github.com/tools/godep) for how to add/update dependencies.

## Unit Test

TODO


Happy hacking!