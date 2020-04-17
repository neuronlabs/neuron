![Neuron Logo](logo_teal.svg)

# Neuron Core [![Go Report Card](https://goreportcard.com/badge/github.com/neuronlabs/neuron-core)](https://goreportcard.com/report/github.com/neuronlabs/neuron-core) [![GoDoc](https://godoc.org/github.com/neuronlabs/neuron-core?status.svg)](https://godoc.org/github.com/neuronlabs/neuron-core) [![Build Status](https://travis-ci.com/neuronlabs/neuron-core.svg?branch=master)](https://travis-ci.com/neuronlabs/neuron-core) ![GitHub](https://img.shields.io/github/license/neuronlabs/neuron-core)


Neuron-core is the golang cloud-native, distributed ORM implementation.

* [What is Neuron?](#what-is-neuron-core)
* [Install](#install)
* [Docs](#docs)
* [Quick Start](#quick-start)
* [Packages](#packages)
* [Design](#design)

## What is Neuron-Core?

Neuron-core is a cloud-ready **Golang** ORM. It's design allows querying multiple related models located on different 
data stores or repositories using a single API.

## Design

Package `neuron` provides golang ORM implementation designed to process microservice repositories. This concept enhanced the requirements for the models mapping and query processor.

The mapped models must have all the field's with [defined type](https://docs.neuronlabs.io/neuron-core/models/structure.html#model-structure). The initial model mapping was based on the [JSONAPI v1.0](https://jsonapi.org/format/#document-resource-objects) model definition. The distributed environment required each [relationship](https://docs.neuronlabs.io/neuron-core/models/relationship.html) to be specified of it's kind and type, just as their foreign keys.

The query processor works as the orchestrator and choreographer. It allows to join the queries where each model's 
use a different database. It also allows distributed transactions [transactions](https://docs.neuronlabs.io/neuron-core/query/transactions.html). 

## Install

`go get github.com/neuronlabs/neuron-core`

## Docs
- Neuron-Core: https://docs.neuronlabs.io/neuron-core
- GoDoc: https://godoc.org/github.com/neuronlabs/neuron-core
- Project Neuron: https://docs.neuronlabs.io

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
    Pets    []*Pet `neuron:"foreign=OwnerID"`
}

// Pet is the model related with the User.
// It is stored in the 'secondary' repository.
type Pet struct {
    ID          int
    Name        string
    OwnerID     int     `neuron:"type=foreign"`
}

// RepositoryName implements RepositoryNamer interface.
func (p *Pet) RepositoryName() string {
    return "secondary"
}
```

* Import repositories and Create, Read or get Default `*config.Controller`
```go
package main

import (
    "context"
    // blank imported repository registers it's factory
    // and the driver.
    _ "github.com/neuronlabs/neuron-postgres"

    "github.com/neuronlabs/neuron-core"
    "github.com/neuronlabs/neuron-core/config"
    "github.com/neuronlabs/neuron-core/log"

)

func main() {
    cfg := config.Default()
    // Provided create config 'cfg' to the Controller method.
    n, err := neuron.New(cfg)
    if err != nil {
        log.Fatal(err)
    } 

    // As the 'neuron-core' allows to use multiple repository for the models
    // we can declare the DefaultRepository within the config. The first 
    // registered repository would be set as the default as well. 
    mainDB := &config.Repository{
        // Currently registered repository 'neuron-pq' has it's driver name: 'pq'.
        DriverName: "postgres",        
        Host: "localhost",   
        Port: 5432,
        Username: "main_db_user",
        Password: "main_db_password",
        DBName: "main",
    }
    if err = n.RegisterRepository("main", mainDB); err != nil {
        log.Fatal(err)
    }

    // We can register and use different repository for other models.
    secondaryDB := &config.Repository{        
        DriverName: "postgres",        
        Host: "172.16.1.10",
        Port: 5432,
        Username: "secondary_user",
        Password: "secondary_password",
        DBName: "secondary",
    }

    // Register secondary repository.
    if err = n.RegisterRepository("secondary", secondaryDB); err != nil {
        log.Fatal(err)
    }

    if err = n.DialAll(context.TODO()); err != nil {
        log.Fatal(err)    
    }
    
    // Register models 
    if err = n.RegisterModels(models.User{}, models.Pet{}); err != nil {
        log.Fatal(err)
    }

    // If necessary migrate the models
    if err = n.MigrateModels(models.User{}, models.Pet{}); err != nil {
        log.Fatal(err)
    }

    // Start application and query models
    users := []*models.User{}
    err = n.Query(&users).        
        .Filter("Name IN","John", "Sam"). // Filter the results by the 'Name' which might be 'John' or 'Sam'.        
        .Sort("-ID"). // Sort the results by user's ids.
        .Include("Pets"). // Include the 'Pets' field for each user in the 'users'.
        .List() // List all the users with the name 'John' or 'Sam' with 'id' ordered in decrease manner.
    if  err != nil {
        log.Fatal(err)
    }
    log.Infof("Queried users: %v", users)
}
```

## Packages

The `neuron-core` is composed of the following packages:

* `query` - used to query the model's repositories.
* `controller` - is the neuron's core, that registers and stores the models and contains configurations required by other packages.
* `mapping` - contains the information about the mapped models their fields and settings
* `config` - contains the configurations structures.
* `class` - contains `github.com/neuronlabs/errors` class instances used by the neuron-core package.
* `repository` - is a package used to store, get and register the repositories nad their factories.
* `log` - is the logging interface for the neuron based applications.

