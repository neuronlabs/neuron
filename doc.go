// Package neuron is the cloud-native, distributed ORM implementation.
//
// It's design allows to use the separate repository for each model, with a possibility
// to have different relationships types between them.
//
// neuron consists of following packages:
// - auth			- defines basic interfaces and structures used for neuron authentication and authorization.
// - codec 			- is a set structures and interfaces used on marshal process.
// - controller 	- defines a structure that keeps and maps all models to related repositories.
// - core 			- contains core neuron service definitions.
// - database		- defines interfaces, functions and structures that allows to execute queries.
// - errors 		- neuron defined errors.
// - log 			- is the neuron service logging interface structure for the neuron based applications.
// - mapping 		- contains the information about the mapped models their fields and settings.
// - query 			- contains structures used to create queries, sort, pagination on base of mapped models.
// - query/filters 	- contains query filters structures and implementations.
// - repository 	- is a package used to store and register the repositories.
// - server			- defines interfaces used as the servers.
//
//	It is also used to get the repository/factory per model. A modular design
//  allows to use and compile only required repositories.
package neuron
