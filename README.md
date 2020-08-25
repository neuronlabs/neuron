![Neuron Logo](logo_teal.svg)

# Neuron [![Go Report Card](https://goreportcard.com/badge/github.com/neuronlabs/neuron)](https://goreportcard.com/report/github.com/neuronlabs/neuron) [![GoDoc](https://godoc.org/github.com/neuronlabs/neuron?status.svg)](https://godoc.org/github.com/neuronlabs/neuron) [![Build Status](https://travis-ci.com/neuronlabs/neuron.svg?branch=master)](https://travis-ci.com/neuronlabs/neuron) ![GitHub](https://img.shields.io/github/license/neuronlabs/neuron)

Neuron is the **golang** framework that uses both code generation and interfaces to create services and microservices. 


* [What is Neuron?](#what-is-neuron)
* [Design](#design)
* [Install](#install)
* [Related Packages](#related-packages)
* [Quick Start](#quick-start)
* [Packages](#packages)

## What is Neuron?

Neuron is a **Golang** framework that enables creation of performance critical, easily scalable services with a focus on the models - not the infrastructure.

It defines a set of structures, interfaces and functions that handles model ORM-like database access, an API server with 
access to its resources, authorization and authentication basics, encoding based marshaling as well as multi version access.        


## Design

The Aim of this project was to provide a fast way of creating performance critical services.

#### Model mapping

With this assumption it was necessary to define interfaces for the models and its database access, a server that 
handles external request and encodes using defined codecs. 

In order to achieve it each model would become scanned and mapped twice - using compile-time code scan and generation and then runtime - dynamic `mapping`. 
This allows getting rid of unnecessary reflection with an ability to know the model on runtime.  

#### Controller

Each model used by the Neuron service is stored within the `controller` and mapped to it's related `repository`.

#### Repository and Queries
A `repository` is a database access interface that allows to execute the queries within specific data storage.

The service uses `query` scopes which defines all required information about the resource request.

#### Database
In order to create and execute queries a package `database` provide set of functions, structures and interfaces that handles
provided `query.Scope` reduce it to the simplest possible form and execute in related model's repository.

This package provides a way to query multiple disjoint repositories and merge its results concurrently using simple query builders.

##### Transactions

This approach required to implement more complicated transaction system.   
Each model may use a different repository, thus an execution of the transaction needs to be divided for each repository.
Each repository is also responsible only for its part of the transaction. 
The whole process of the transaction is being orchestrated by the `database.Tx`.
     

## Install

`go get github.com/neuronlabs/neuron`

## Related packages

##### [Code Generation](https://github.com/neuronlabs/neurogonesis)

Neuron requires the model to implement multiple interfaces. In order to make this process faster and easier a code generation 
tool `neurogonesis` - [https://github.com/neuronlabs/neurogonesis](https://github.com/neuronlabs/neurogonesis) is implemented. 

It contains a CLI tool that might be used as `go:generate` to scan golang code and implement all required interfaces for your models.
See to check more details about `go:generate` - [https://blog.golang.org/generate](https://blog.golang.org/generate).

It could also generate `Collections` - which are the model specific non-interface  database query builder and composer.     

##### [Extensions](https://github.com/neuronlabs/neuron-extensions)

Packages in the `Neuron` repository provides mostly basic structures, interfaces and functions, in order not to limit it to specific 
implementations. 

Repository [https://github.com/neuronlabs/neuron-extensions](https://github.com/neuronlabs/neuron-extensions) - 
contains implementations for specific codecs, servers, repositories and authorization/authentication services.

All of these could be used as modules that creates your neuron service faster and better.

##### [Examples](https://github.com/neuronlabs/neuron-examples)

How to get known with the Neuron better? Visit repository [https://github.com/neuronlabs/neuron-examples](https://github.com/neuronlabs/neuron-examples) - that shows examples on how the services could be created,
for specific usages. 


## Quick Start

##### Define the models

`models.go`
```go
package main

// Define go:generate generators and use it every time you change any model - using go:generate.
//go:generate neurogonesis models --format=goimports --single-file .
//go:generate neurogonesis collections --format=goimports  --single-file .

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
    Owner       *User   
    OwnerID     int
}

```

##### Create repository and server

`repository.go`

```go
package main

import (
	"os"

	"github.com/neuronlabs/neuron-extensions/repository/postgres"
	"github.com/neuronlabs/neuron/log"
	"github.com/neuronlabs/neuron/repository"
)

const (
    // Environment variable that contains postgres repository access URI for the default repository.
    envDefaultPostgres  = "NEURON_DEFAULT_POSTGRES"
    // Environment variable that contains postgres repository access URI for the pets. 
    envPetsPostgres     = "NEURON_PETS_POSTGRES" 
)

// defaultRepository gets the default postgres repository configuration for the service.
func defaultRepository() *postgres.Postgres {
	uriCredentials, ok := os.LookupEnv(envDefaultPostgres)
	if !ok {
		log.Fatalf("required environment variable: %s not found", envDefaultPostgres)
	}
	return postgres.New(repository.WithURI(uriCredentials))
}

// petsRepository gets the repository configuration for the Comments model.
func petsRepository() *postgres.Postgres {
	uriCredentials, ok := os.LookupEnv(envPetsPostgres)
	if !ok {
		log.Fatalf("required environment variable %s not found", envPetsPostgres)
	}
	return postgres.New(repository.WithURI(uriCredentials))
}
```

##### Server
```go
package main

import (
	"github.com/neuronlabs/neuron-extensions/server/http"
	"github.com/neuronlabs/neuron-extensions/server/http/api/jsonapi"
	"github.com/neuronlabs/neuron-extensions/server/http/middleware"
)

func getJsonAPI() *jsonapi.API {
	// Define new json:api specification API.
	api := jsonapi.New(
		jsonapi.WithMiddlewares(middleware.ResponseWriter(-1), middleware.LogRequest),
		// Set path prefix to '/v1/api/
		jsonapi.WithPathPrefix("/v1/api"),
		// Set the strict unmarshal flag which disallows unknown model fields in the documents.
		jsonapi.WithStrictUnmarshal(),
		// Set the models with default api handler.
		jsonapi.WithDefaultHandlerModels(&User{}, &Pet{}),		
	)
	return api
}

func getServer() *http.Server {
	// Create api based on json:api specification.
	jsonAPI := getJsonAPI()
	// Create new http server.
	s := http.New(
		// Mount json:api with the models.
		http.WithAPI(jsonAPI),
		// Set the listening port.
		http.WithPort(8080),
	)
	return s
}

```

##### Main service
`main.go` 
 
```go
package main


import (
	"context"
	"net/http"

	"github.com/neuronlabs/neuron"
	"github.com/neuronlabs/neuron/log"

	serverLogs "github.com/neuronlabs/neuron-extensions/server/http/log"		
)

func main() {
	log.NewDefault()
	n := neuron.New(
		// Set the http server with json:api into service server.
		neuron.Server(getServer()),
		// Define the default repository name for all models without repository name specified.
		neuron.DefaultRepository(defaultRepository()),
		// Set the custom repository for the comments model.
		neuron.RepositoryModels(petsRepository(), &Pet{}),
		// Set the models in the service - Neuron_Models are defined by the go:generate models
		neuron.Models(Neuron_Models...),
		// Migrate models into service - this would create database definitions for the provided models.
		neuron.MigrateModels(Neuron_Models...),
		// Initialize model collections - Neuron_Collections 
		neuron.Collections(Neuron_Collections...),
	)
	log.SetLevel(log.LevelDebug3)
	serverLogs.SetLevel(log.LevelDebug3)

	ctx := context.Background()
    // Initialize the service
	if err := n.Initialize(ctx); err != nil {
		log.Fatalf("Initialize failed: %v", err)
	}

	// List all endpoints defined in the server and print their paths.
	for _, endpoint := range n.Server.GetEndpoints() {
		log.Infof("Endpoint [%s] %s", endpoint.HTTPMethod, endpoint.Path)
	}
	if err := n.Run(ctx); err != nil && err != http.ErrServerClosed {
		log.Errorf("Running neuron service failed: %s", err)
	}
}
```

## Packages

The `neuron` is composed of the following packages:

* `auth`			- defines basic interfaces and structures used for neuron authentication and authorization.
* `codec` 			- is a set structures and interfaces used on marshal process.
* `core` 	        - defines a structure that keeps and maps all models to related repositories.
* `database`		- defines interfaces, functions and structures that allows to execute queries.
* `errors` 		    - neuron defined errors.
* `log` 			- is the neuron service logging interface structure for the neuron based applications.
* `filestore`       - defines common interface for the file stores.
* `mapping` 		- contains the information about the mapped models their fields and settings.
* `query` 			- contains structures used to create queries, sort, pagination on base of mapped models.
* `query/filters` 	- contains query filters structures and implementations.
* `repository` 	    - is a package used to store and register the repositories.
* `server`			- defines interfaces used as the servers.
* `service`			- contains core neuron service definitions.
* `store`           - key-value stores structures and implementations.


## Docs
- neuron: https://docs.neuronlabs.io/neuron
- GoDoc: https://godoc.org/github.com/neuronlabs/neuron
- Project Neuron: https://docs.neuronlabs.io
