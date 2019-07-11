package class

// MjrRepository is the major error classification
// related with the repositories.
var MjrRepository Major

func registerRepositoryClasses() {
	MjrRepository = MustRegisterMajor("Repository", "repositories related errors")

	registerRepositoryUnavailable()
	registerRepositoryAuth()
	registerRepositoryConnection()
	registerRepositoryInterface()
	registerRepositoryFactory()
	registerRepositoryModel()
	registerRepostioryUnmappedError()
	registerRepositoryConfig()
	registerRepositoryReplica()
}

/**

Repository Unavailable

*/

var (
	// MnrRepositoryUnavailable is a 'MjrRepository' minor error classification
	// for unavailable repository access.
	MnrRepositoryUnavailable Minor

	// RepositoryUnavailableInsufficientResources is the 'MjrRepository', 'MnrRepositoryUnavailable' error classification
	// when the repository has insufficitent system resources to run.
	RepositoryUnavailableInsufficientResources Class

	// RepositoryUnavailableProgramLimit is the 'MjrRepository', 'MnrRepositoryUnavailable' error classification
	// when the repository reached program limit - i.e. tried to extract too many columns at once.
	RepositoryUnavailableProgramLimit Class

	// RepositoryUnavailableShutdown is the 'MjrRepository', 'MnrRepositoryUnavailable' error classification
	// when the repository is actually shutting down.
	RepositoryUnavailableShutdown Class
)

func registerRepositoryUnavailable() {
	MnrRepositoryUnavailable = MjrRepository.MustRegisterMinor("Unavailable", "current repository access is unavailable")

	RepositoryUnavailableInsufficientResources = MnrRepositoryUnavailable.MustRegisterIndex("Insufficient Resources",
		"current repository is out of possible resources").Class()
	RepositoryUnavailableProgramLimit = MnrRepositoryUnavailable.MustRegisterIndex("Program Limit",
		"reached program limit - i.e. too many fields to return").Class()
	RepositoryUnavailableShutdown = MnrRepositoryUnavailable.MustRegisterIndex("Shutdown", "repository is currently shutting down").Class()
}

/**

Repository Authorization / Authentication

*/

var (
	// MnrRepositoryAuth is the 'MjrRepository' minor error classification
	// related with Authorization or Authentication issues.
	MnrRepositoryAuth Minor

	// RepositoryAuthPrivileges is the 'MjrRepository', 'MnrRepositoryAuth' error classification
	// for insufficient authorization privileges issues.
	RepositoryAuthPrivileges Class
)

func registerRepositoryAuth() {
	MnrRepositoryAuth = MjrRepository.MustRegisterMinor("Authorization", "repositories authorization issues")
	RepositoryAuthPrivileges = MnrRepositoryAuth.MustRegisterIndex("Privileges", "insufficient authorization privileges issues").Class()
}

/**

Repository Connection

*/

var (
	// MnrRepositoryConnection is the 'MjrRepository' minor error classification
	// for the repository connection issues.
	MnrRepositoryConnection Minor

	// RepositoryConnection is the 'MjrRepository', 'MnrRepositoryConnection'
	// error classification related with repository connection.
	RepositoryConnection Class

	// RepositoryConnectionTimedOut is the 'MjrRepository', 'MnrRepositoryConnection'
	// error classification related with timed out connection.
	RepositoryConnectionTimedOut Class

	// RepositoryConnectionURI is the 'MjrRepository', 'MnrRepositoryConnection' error classification
	// related with invalid connection URI for the repository.
	RepositoryConnectionURI Class

	// RepositoryConnectionSSL is the 'MjrRepository', 'MnrRepositoryConnection' error classification
	// related with SSL for the repository connections.
	RepositoryConnectionSSL Class
)

func registerRepositoryConnection() {
	MnrRepositoryConnection = MjrRepository.MustRegisterMinor("Connection", "repository connection issues")

	RepositoryConnection = MustNewMinorClass(MnrRepositoryConnection)
	RepositoryConnectionTimedOut = MnrRepositoryConnection.MustRegisterIndex("Timed Out", "repository connection timed out").Class()
	RepositoryConnectionURI = MnrRepositoryConnection.MustRegisterIndex("URI", "invalid URI provided for the repository connection").Class()
	RepositoryConnectionSSL = MnrRepositoryConnection.MustRegisterIndex("SSL", "SSL failed").Class()

}

/**

Repository Interface

*/

var (
	// MnrRepositoryNotImplements is the 'MjrRepository' minor error classification
	// for the repository interfaces.
	MnrRepositoryNotImplements Minor

	// RepositoryNotImplementsTransactioner is the 'MjrRepository', 'MnrRepositoryInterface' error classification
	// for errors when the repository doesn't implement transactioner interfaces.
	RepositoryNotImplementsTransactioner Class

	// RepositoryNotImplementsCreator is the 'MjrRepository', 'MnrRepositoryInterface' error classification
	// for errors when the repository doesn't implement query.Creator interfaces.
	RepositoryNotImplementsCreator Class

	// RepositoryNotImplementsDeleter is the 'MjrRepository', 'MnrRepositoryInterface' error classification
	// for errors when the repository doesn't implement query.Deleter interfaces.
	RepositoryNotImplementsDeleter Class

	// RepositoryNotImplementsPatcher is the 'MjrRepository', 'MnrRepositoryInterface' error classification
	// for errors when the repository doesn't implement query.Patcher interfaces.
	RepositoryNotImplementsPatcher Class

	// RepositoryNotImplementsLister is the 'MjrRepository', 'MnrRepositoryInterface' error classification
	// for errors when the repository doesn't implement query.Lister interfaces.
	RepositoryNotImplementsLister Class

	// RepositoryNotImplementsGetter is the 'MjrRepository', 'MnrRepositoryInterface' error classification
	// for errors when the repository doesn't implement getter interfaces.
	RepositoryNotImplementsGetter Class
)

func registerRepositoryInterface() {
	MnrRepositoryNotImplements = MjrRepository.MustRegisterMinor("NotImplements", "repository implenting interface issues")

	RepositoryNotImplementsCreator = MnrRepositoryNotImplements.MustRegisterIndex("Creator", "repository doesn't implement creater").Class()
	RepositoryNotImplementsDeleter = MnrRepositoryNotImplements.MustRegisterIndex("Deleter", "repository doesn't implement deleter").Class()
	RepositoryNotImplementsGetter = MnrRepositoryNotImplements.MustRegisterIndex("Getter", "repository doesn't implement getter").Class()
	RepositoryNotImplementsLister = MnrRepositoryNotImplements.MustRegisterIndex("Lister", "repository doesn't implement lister").Class()
	RepositoryNotImplementsPatcher = MnrRepositoryNotImplements.MustRegisterIndex("Patcher", "repository doesn't implement patcher").Class()
	RepositoryNotImplementsTransactioner = MnrRepositoryNotImplements.MustRegisterIndex("Transactioner", "repository doesn't implement transactioner").Class()
}

var (
	// MnrRepositoryNotFound is the 'MjrRepository' minor error classification
	// when the repository is not found.
	MnrRepositoryNotFound Minor

	// RepositoryNotFound is the 'MjrRepository', 'MnrRepositoryNotFound' errors classification
	// when the repository for model is not found.
	RepositoryNotFound Class
)

func registerRepositoryNotFound() {
	MnrRepositoryNotFound = MjrRepository.MustRegisterMinor("Not Found", "repository not found")

	RepositoryNotFound = MustNewMinorClass(MnrRepositoryNotFound)
}

var (
	// MnrRepositoryFactory is the 'MjrRepository' minor error classification
	// for repository factories issues.
	MnrRepositoryFactory Minor

	// RepositoryFactoryNotFound is the 'MjrRepository', 'MnrRepositoryFactory' error classification
	// for not found repository factory.
	RepositoryFactoryNotFound Class

	// RepositoryFactoryAlreadyRegistered is the 'MjrRepository', 'MnrRepositoryFactory' error classification
	// for the factories that were already registered.
	RepositoryFactoryAlreadyRegistered Class
)

func registerRepositoryFactory() {
	MnrRepositoryFactory = MjrRepository.MustRegisterMinor("Factory", "isseus with the repository factory")

	RepositoryFactoryNotFound = MnrRepositoryFactory.MustRegisterIndex("Not Found", "repository factory not found").Class()
	RepositoryFactoryAlreadyRegistered = MnrRepositoryFactory.MustRegisterIndex("Already Registered", "repository factory already registered").Class()
}

// MnrRepositoryModel is the minor error classification for the repository model issues.
var MnrRepositoryModel Minor

var (
	// RepositoryModelReservedName is the 'MjrRepository', 'MnrRepositoryModel' error classification
	// for reserved names in the repositories.
	RepositoryModelReservedName Class

	// RepositoryModelTags is the 'MjrRepository', 'MnrRepositoryModel' error classification
	// for invalid repository specific field tags.
	RepositoryModelTags Class
)

func registerRepositoryModel() {
	MnrRepositoryModel = MjrRepository.MustRegisterMinor("Model", "repository specific model issues")

	RepositoryModelReservedName = MnrRepositoryModel.MustRegisterIndex("Reserved Name", "reserved name violation").Class()
	RepositoryModelTags = MnrRepositoryModel.MustRegisterIndex("Tags", "repository specific tags issues").Class()
}

// RepositoryUnmappedError is the error 'MjrRepository' classification used for unmapped
// repository specific error classes, types or codes.
var RepositoryUnmappedError Class

func registerRepostioryUnmappedError() {
	RepositoryUnmappedError = MustNewMinorClass(MjrRepository.MustRegisterMinor("Unmapped Error", "unmapped repository errors"))
}

var (
	// MnrRepositoryConfig is the 'MjrRepository' error classification for repository configurations
	MnrRepositoryConfig Minor

	// RepositoryConfigAlreadyRegistered is the 'MjrRepository', 'MnrRepositoryConfig' error classification
	// for already registered repository configurations.
	RepositoryConfigAlreadyRegistered Class

	// RepositoryConfigInvalid is the 'MjrRepository', 'MnrRepositoryConfig' error classification
	// for invalid repository configuration.
	RepositoryConfigInvalid Class
)

func registerRepositoryConfig() {
	MnrRepositoryConfig = MjrRepository.MustRegisterMinor("Config", "repository configurations")

	RepositoryConfigAlreadyRegistered = MnrRepositoryConfig.MustRegisterIndex("Already Registered", "repository configuration already exists").Class()
	RepositoryConfigInvalid = MnrRepositoryConfig.MustRegisterIndex("Invalid", "invalid repository configuration").Class()
}

var (
	// MnrRepositoryReplica is the 'MjrRepository' minor error classification
	// for replicas errors.
	MnrRepositoryReplica Minor

	// RepositoryReplicaSetNotFound is the 'MjrRepository', 'MnrRepositoryReplica' error classification
	// where replica set is not found.
	RepositoryReplicaSetNotFound Class
	// RepositoryReplicaShardNotFound is the 'MjrRepository', 'MnrRepositoryReplica' error classification
	// where replica shard is not found.
	RepositoryReplicaShardNotFound Class
	// RepositoryReplicaNodeNotFound is the 'MjrRepository', 'MnrRepositoryReplica' error classification
	// where replica shard is not found.
	RepositoryReplicaNodeNotFound Class

	// RepositoryReplica is the 'MjrRepository', 'MnrRepositoryReplica' error classification for repository replicas.
	RepositoryReplica Class
)

func registerRepositoryReplica() {
	MnrRepositoryReplica = MjrRepository.MustRegisterMinor("Replica", "repository replicas")

	RepositoryReplicaSetNotFound = MnrRepositoryReplica.MustRegisterIndex("Set Not Found", "replica set is not found").Class()
	RepositoryReplicaShardNotFound = MnrRepositoryReplica.MustRegisterIndex("Shard Not Found", "replica shard not found").Class()
	RepositoryReplicaNodeNotFound = MnrRepositoryReplica.MustRegisterIndex("Node Not Found", "replica node not found").Class()
	RepositoryReplica = MustNewMinorClass(MnrRepositoryReplica)
}
