# Service Label

The `servicelabel` is a small Core Agent Plugin which other plugins can use to
obtain the microservice label, i.e. the string used to identify the particular VNF.
The label is primarily used to prefix keys in ETCD datastore so that the configurations
of different VNFs do not get mixed up.

**API**

The API cannot be simpler. Plugin can obtain the microservice label using the function `GetAgentLabel()`,
which really just returns a string value already obtained during the `servicelabel` initialization phase.

**Example**

Example of retrieving and using the microservice label:
```
plugin.Label = servicelabel.GetAgentLabel()
dbw.Watch(dataChan, cfg.SomeConfigKeyPrefix(plugin.Label))
```
