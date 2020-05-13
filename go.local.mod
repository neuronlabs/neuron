module github.com/neuronlabs/neuron

replace (
    github.com/neuronlabs/neuron/errors => ./errors
)

go 1.11

require (
	github.com/google/uuid v1.1.1
	github.com/kr/pretty v0.1.0 // indirect
	github.com/neuronlabs/inflection v1.0.1
	github.com/neuronlabs/neuron/errors v0.0.0-20200511120829-fff1f8cf09c7
	github.com/neuronlabs/strcase v1.0.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.3 // indirect
)
