# Task

[![Build Status](https://travis-ci.org/denderello/task.svg?branch=master)](https://travis-ci.org/denderello/task)
[![Go Report Card](https://goreportcard.com/badge/denderello/task "Go Report Card")](https://goreportcard.com/report/denderello/task)

`task` is a simple Go library to execute tasks after each other. When one of the
tasks fails the queue will be walked backwards and each task gets the
opportunity to run rollback actions of the actions performed beforehand.

## Installation
To install, simply execute:
```
go get github.com/denderello/task
```

## Example
You can find a simple example in [examples/bool_tasks.go](examples/bool_tasks.go) which you can run like
this:
```
go run examples/bool_tasks.go
```
