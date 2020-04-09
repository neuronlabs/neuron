# Controller creation scenarios

## Initialization Flow

- Create Controller `controller.New(cfg)`
- Register Repositories `c.RegisterRepository(repositoryConfig)`
- Dial with repositories to check if there is no problem. `c.Dial(ctx)`
- Register Models - registers all models in `neuron` and related repositories `c.RegisterModels(models...)`
- (Optional) Migrate models (i.e. create tables, collections etc.) `c.MigrateModels(models...)`
- Run application
- On close, SIGINT, SIGTERM close all repository connections - `c.Close(ctx)`. 

## Scenarios with Default Controller:

On initialization the default controller doesn't contain any repository
but have threadsafe processor.


## Scenario with Custom Controller:

The controller is created with some config for creation.

## Creation Rules

- processor check on creation:
    - check if processor is set
    - otherwise check the ProcessorName and map it with registered processors
- repository check on creation:
    - namerFunc can't be empty
    - logger should be set if level provided
    - set validation aliases

## Register Model rules

- If model have repository defined:
    - the name must be defined
    - register if not exists
- If model have repository name defined:
    - the repository must be registered
    - map the model to this repository
- If model have no repository nor repository name defined:
    - the default repository must be registered
    - map the model to the default repository

