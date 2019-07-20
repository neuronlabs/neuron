// Package ncore is the cloud-native, distributed ORM implementation.
// It's design allows to use the separate repository
// for each model, with a possiblity to have different
// relationships types between them.
// It consits the following packages:
// - ncore - (Neuron Core) the root package that gives easy access to all subpackages.
// - common - contains common variables and constants for neuron derivates.
// - controller - is the neuron's core, that registers and stores the models and
// contains configurations required by other packages.
// - config - contains the configurations for all packages.
// - query - used to query the model's repositories.
// - mapping - contains the information about the mapped models their fields and settings.
// - errors - used as a default error package for the neuron packages.
// - errors/class - contains errors classification system for the neuron packages.
// - encoding/jsonapi - allows to marshal and unmarshal the
//	model's, and queries by the 'https://jsonapi.org/' specification
// - log - is the logging interface for the neuron based applications.
// - i18n - is the neuron based application supported internationalization.
// - repository - is a package used to store and register the repositories.
//	It is also used to get the repository/factory per model. A modular design
//  allows to use and compile only required repositories.
package ncore
