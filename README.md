# neuron-core

Neuron-core is the golang cloud-native, distributed ORM implementation.

[![Go Report Card](https://goreportcard.com/badge/github.com/neuronlabs/neuron-core)](https://goreportcard.com/report/github.com/neuronlabs/neuron-core)
[![GoDoc](https://godoc.org/github.com/neuronlabs/neuron-core?status.svg)](https://godoc.org/github.com/neuronlabs/neuron-core)

* [What is Neuron?](#what-is-neuron)
* [Install](#install)
* [Docs](#docs)
* [Quick Start](#quick-start)

## What is Neuron-Core?
Neuron-core is a cloud-ready **Golang** ORM. It's design allows to
query multiple related models located on different datastores/repositories.

## Install

`go get -u github.com/neuronlabs/neuron-core`

## Docs

- Neuron-Core: `https://neuronlabs.github.io/neuron-core`
- GoDoc: `https://godoc.org/github.com/neuronlabs/neuron-core`
- Addons and Repositories: `https://neuronlabs.github.io/`

## Quick Start

* Define the models
```go
package models

// User is the model that is stored on the 'main' repository.
// It is related to multiple 'Pet' models.
type User struct {
    ID      int
    Name    string
    Surname string
    Pets []*Pet `neuron:"type=relation;foreign=OwnerID"`
}

// Pet is the model related with the User.
// It is stored in the 'secondary' repository.
type Pet struct {
    ID      int
    Name    string
    OwnerID int `neuron:"type=foreign"`
}

// RepositoryName implements RepositoryNamer interface.
func (p *Pet) RepositoryName() string {
    return "secondary"
}
```

* Import repositories and Create, Read or get Default `*config.Controller`
```go
package main

import(
    // blank imported repository registers it's factory
    // and the driver.
    _ "github.com/neuronlabs/neuron-pq"
)

import (
    "github.com/neuronlabs/neuron-core/config"
    "github.com/neuronlabs/neuron-core"
)

func main() {
    cfg := config.ReadDefaultConfig()
    // By setting the LogLevel the default logger would be used.
    cfg.LogLevel = "debug"    
...    
```
* Create the `*controller.Controller` and register repositories.
```go
    // Provided create config 'cfg' to the Controller method.
    c := ncore.Controller(cfg)

    // As the 'neuron-core' allows to use multiple repository for the models
    // we can declare the DefaultRepository within the config. The first 
    // registered repository would be set as the default as well. 
    mainDB := &config.Repository{
        // Currently registered repository 'neuron-pq' has it's driver name: 'pq'.
        DriverName: "pq",        
        Host: "localhost",   
        Port: "5432",
        Username: "main_db_user",
        Password: "main_db_password",
        DBName: "main",
    }
    if err := c.RegisterRepository("main", mainDB); err != nil {
        panic(err)
    }

    // We can register and use different repository for other models.
    secondaryDB := &config.Repository{        
        DriverName: "pq",        
        Host: "172.16.1.10",
        Port: "5432",
        Username: "secondary_user",
        Password: "secondary_password",
        DBName: "secondary",
    }

    // Register secondary repository.
    if err := c.RegisterRepository("secondary", secondaryDB); err != nil {
        panic(err)
    }
```

* Register models 
```go
    if err := c.RegisterModels(models.User{}, models.Pet{}); err != nil {
        panic(err)
    }
```
* Query registered models
```go
    users := []*User{}
    
    s := ncore.MustQueryC(c, &users)
    // the query scope may be filtered
    s.AddStringFilter("filter[users][name][$in]","John", "Sam")
    // it might also be sorted
    s.SortBy("-id")
    
    // list all the users with the name 'John' or 'Sam' with 'id' ordered 
    // descending.
    if err = s.List(); err != nil {
        panic(err)
    }
```



