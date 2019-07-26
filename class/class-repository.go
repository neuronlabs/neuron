package class

import (
	"github.com/neuronlabs/errors"
)

// MjrRepository is the major error classification
// related with the repositories.
var MjrRepository errors.Major

func registerRepositoryClasses() {
	MjrRepository = errors.NewMajor()

	registerRepositoryUnavailable()
	registerRepositoryAuth()
	registerRepositoryConnection()
	registerRepositoryInterface()
	registerRepositoryFactory()
	registerRepositoryModel()
	registerRepostioryUnmappedError()
	registerRepositoryConfig()
	registerRepositoryReplica()
	registerRepositoryNotFound()
}

/**

Repository Unavailable

*/

var (
	// MnrRepositoryUnavailable is a 'MjrRepository' minor error classification
	// for unavailable repository access.
	MnrRepositoryUnavailable errors.Minor

	// RepositoryUnavailableInsufficientResources is the 'MjrRepository', 'MnrRepositoryUnavailable' error classification
	// when the repository has insufficitent system resources to run.
	RepositoryUnavailableInsufficientResources errors.Class

	// RepositoryUnavailableProgramLimit is the 'MjrRepository', 'MnrRepositoryUnavailable' error classification
	// when the repository reached program limit - i.e. tried to extract too many columns at once.
	RepositoryUnavailableProgramLimit errors.Class

	// RepositoryUnavailableShutdown is the 'MjrRepository', 'MnrRepositoryUnavailable' error classification
	// when the repository is actually shutting down.
	RepositoryUnavailableShutdown errors.Class
)

func registerRepositoryUnavailable() {
	MnrRepositoryUnavailable = errors.MustNewMinor(MjrRepository)

	mjr, mnr := MjrRepository, MnrRepositoryUnavailable
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	RepositoryUnavailableInsufficientResources = newClass()
	RepositoryUnavailableProgramLimit = newClass()
	RepositoryUnavailableShutdown = newClass()
}

/**

Repository Authorization / Authentication

*/

var (
	// MnrRepositoryAuth is the 'MjrRepository' minor error classification
	// related with Authorization or Authentication issues.
	MnrRepositoryAuth errors.Minor

	// RepositoryAuthPrivileges is the 'MjrRepository', 'MnrRepositoryAuth' error classification
	// for insufficient authorization privileges issues.
	RepositoryAuthPrivileges errors.Class
)

func registerRepositoryAuth() {
	MnrRepositoryAuth = errors.MustNewMinor(MjrRepository)
	RepositoryAuthPrivileges = errors.MustNewClass(MjrRepository, MnrRepositoryAuth, errors.MustNewIndex(MjrRepository, MnrRepositoryAuth))
}

/**

Repository Connection

*/

var (
	// MnrRepositoryConnection is the 'MjrRepository' minor error classification
	// for the repository connection issues.
	MnrRepositoryConnection errors.Minor

	// RepositoryConnection is the 'MjrRepository', 'MnrRepositoryConnection'
	// error classification related with repository connection.
	RepositoryConnection errors.Class

	// RepositoryConnectionTimedOut is the 'MjrRepository', 'MnrRepositoryConnection'
	// error classification related with timed out connection.
	RepositoryConnectionTimedOut errors.Class

	// RepositoryConnectionURI is the 'MjrRepository', 'MnrRepositoryConnection' error classification
	// related with invalid connection URI for the repository.
	RepositoryConnectionURI errors.Class

	// RepositoryConnectionSSL is the 'MjrRepository', 'MnrRepositoryConnection' error classification
	// related with SSL for the repository connections.
	RepositoryConnectionSSL errors.Class
)

func registerRepositoryConnection() {
	MnrRepositoryConnection = errors.MustNewMinor(MjrRepository)

	mjr, mnr := MjrRepository, MnrRepositoryConnection
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	RepositoryConnection = errors.MustNewMinorClass(mjr, mnr)
	RepositoryConnectionTimedOut = newClass()
	RepositoryConnectionURI = newClass()
	RepositoryConnectionSSL = newClass()

}

/**

Repository Interface

*/

var (
	// MnrRepositoryNotImplements is the 'MjrRepository' minor error classification
	// for the repository interfaces.
	MnrRepositoryNotImplements errors.Minor

	// RepositoryNotImplementsTransactioner is the 'MjrRepository', 'MnrRepositoryInterface' error classification
	// for errors when the repository doesn't implement transactioner interfaces.
	RepositoryNotImplementsTransactioner errors.Class

	// RepositoryNotImplementsCreator is the 'MjrRepository', 'MnrRepositoryInterface' error classification
	// for errors when the repository doesn't implement query.Creator interfaces.
	RepositoryNotImplementsCreator errors.Class

	// RepositoryNotImplementsDeleter is the 'MjrRepository', 'MnrRepositoryInterface' error classification
	// for errors when the repository doesn't implement query.Deleter interfaces.
	RepositoryNotImplementsDeleter errors.Class

	// RepositoryNotImplementsPatcher is the 'MjrRepository', 'MnrRepositoryInterface' error classification
	// for errors when the repository doesn't implement query.Patcher interfaces.
	RepositoryNotImplementsPatcher errors.Class

	// RepositoryNotImplementsLister is the 'MjrRepository', 'MnrRepositoryInterface' error classification
	// for errors when the repository doesn't implement query.Lister interfaces.
	RepositoryNotImplementsLister errors.Class

	// RepositoryNotImplementsGetter is the 'MjrRepository', 'MnrRepositoryInterface' error classification
	// for errors when the repository doesn't implement getter interfaces.
	RepositoryNotImplementsGetter errors.Class
)

func registerRepositoryInterface() {
	MnrRepositoryNotImplements = errors.MustNewMinor(MjrRepository)

	mjr, mnr := MjrRepository, MnrRepositoryNotImplements
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	RepositoryNotImplementsCreator = newClass()
	RepositoryNotImplementsDeleter = newClass()
	RepositoryNotImplementsGetter = newClass()
	RepositoryNotImplementsLister = newClass()
	RepositoryNotImplementsPatcher = newClass()
	RepositoryNotImplementsTransactioner = newClass()
}

var (
	// MnrRepositoryNotFound is the 'MjrRepository' minor error classification
	// when the repository is not found.
	MnrRepositoryNotFound errors.Minor

	// RepositoryNotFound is the 'MjrRepository', 'MnrRepositoryNotFound' errors classification
	// when the repository for model is not found.
	RepositoryNotFound errors.Class
)

func registerRepositoryNotFound() {
	MnrRepositoryNotFound = errors.MustNewMinor(MjrRepository)

	RepositoryNotFound = errors.MustNewMinorClass(MjrRepository, MnrRepositoryNotFound)
}

var (
	// MnrRepositoryFactory is the 'MjrRepository' minor error classification
	// for repository factories issues.
	MnrRepositoryFactory errors.Minor

	// RepositoryFactoryNotFound is the 'MjrRepository', 'MnrRepositoryFactory' error classification
	// for not found repository factory.
	RepositoryFactoryNotFound errors.Class

	// RepositoryFactoryAlreadyRegistered is the 'MjrRepository', 'MnrRepositoryFactory' error classification
	// for the factories that were already registered.
	RepositoryFactoryAlreadyRegistered errors.Class
)

func registerRepositoryFactory() {
	MnrRepositoryFactory = errors.MustNewMinor(MjrRepository)

	mjr, mnr := MjrRepository, MnrRepositoryFactory
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	RepositoryFactoryNotFound = newClass()
	RepositoryFactoryAlreadyRegistered = newClass()
}

// MnrRepositoryModel is the minor error classification for the repository model issues.
var MnrRepositoryModel errors.Minor

var (
	// RepositoryModelReservedName is the 'MjrRepository', 'MnrRepositoryModel' error classification
	// for reserved names in the repositories.
	RepositoryModelReservedName errors.Class

	// RepositoryModelTags is the 'MjrRepository', 'MnrRepositoryModel' error classification
	// for invalid repository specific field tags.
	RepositoryModelTags errors.Class
)

func registerRepositoryModel() {
	MnrRepositoryModel = errors.MustNewMinor(MjrRepository)

	mjr, mnr := MjrRepository, MnrRepositoryModel
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	RepositoryModelReservedName = newClass()
	RepositoryModelTags = newClass()
}

// RepositoryUnmappedError is the error 'MjrRepository' classification used for unmapped
// repository specific error classes, types or codes.
var RepositoryUnmappedError errors.Class

func registerRepostioryUnmappedError() {
	RepositoryUnmappedError = errors.MustNewMinorClass(MjrRepository, errors.MustNewMinor(MjrRepository))
}

var (
	// MnrRepositoryConfig is the 'MjrRepository' error classification for repository configurations
	MnrRepositoryConfig errors.Minor

	// RepositoryConfigAlreadyRegistered is the 'MjrRepository', 'MnrRepositoryConfig' error classification
	// for already registered repository configurations.
	RepositoryConfigAlreadyRegistered errors.Class

	// RepositoryConfigInvalid is the 'MjrRepository', 'MnrRepositoryConfig' error classification
	// for invalid repository configuration.
	RepositoryConfigInvalid errors.Class
)

func registerRepositoryConfig() {
	MnrRepositoryConfig = errors.MustNewMinor(MjrRepository)

	mjr, mnr := MjrRepository, MnrRepositoryConfig
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	RepositoryConfigAlreadyRegistered = newClass()
	RepositoryConfigInvalid = newClass()
}

var (
	// MnrRepositoryReplica is the 'MjrRepository' minor error classification
	// for replicas errors.
	MnrRepositoryReplica errors.Minor

	// RepositoryReplicaSetNotFound is the 'MjrRepository', 'MnrRepositoryReplica' error classification
	// where replica set is not found.
	RepositoryReplicaSetNotFound errors.Class
	// RepositoryReplicaShardNotFound is the 'MjrRepository', 'MnrRepositoryReplica' error classification
	// where replica shard is not found.
	RepositoryReplicaShardNotFound errors.Class
	// RepositoryReplicaNodeNotFound is the 'MjrRepository', 'MnrRepositoryReplica' error classification
	// where replica shard is not found.
	RepositoryReplicaNodeNotFound errors.Class

	// RepositoryReplica is the 'MjrRepository', 'MnrRepositoryReplica' error classification for repository replicas.
	RepositoryReplica errors.Class
)

func registerRepositoryReplica() {
	MnrRepositoryReplica = errors.MustNewMinor(MjrRepository)

	mjr, mnr := MjrRepository, MnrRepositoryReplica
	newClass := func() errors.Class {
		return errors.MustNewClass(mjr, mnr, errors.MustNewIndex(mjr, mnr))
	}

	RepositoryReplicaSetNotFound = newClass()
	RepositoryReplicaShardNotFound = newClass()
	RepositoryReplicaNodeNotFound = newClass()
	RepositoryReplica = errors.MustNewMinorClass(mjr, mnr)
}
