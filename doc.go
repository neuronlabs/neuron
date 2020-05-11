// Package neuron is the cloud-native, distributed ORM implementation.
//
// It's design allows to use the separate repository for each model, with a possibility
// to have different relationships types between them.
//
// neuron consists of following packages:
// - config - contains the configurations for all packages.
// - controller - is the neuron's core, that registers and stores the models and
// 	 			  contains configurations required by other packages.
// - errors - classified errors.
// - query - used to create queries, filters, sort, pagination on base of mapped models.
// - mapping - contains the information about the mapped models their fields and settings.
// - log - is the neuron service logging interface structure for the neuron based applications.
// - repository - is a package used to store and register the repositories.
//
//	It is also used to get the repository/factory per model. A modular design
//  allows to use and compile only required repositories.
package neuron
