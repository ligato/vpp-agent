# Client v1

Client v1 (i.e. the first version) defines API that allows to manage configuration of default plugins and linux plugin.
The way of configuration transport from a API user to the plugins is abstracted.
The API calls can be split into two groups:
 - **resync** applies the given configuration. If a configuration exists it is replaced. (The name is abbreviation
  of *resynchronization* that is required if the previous config is not known.) 
 - **data change** allows to deliver incremental changes of a configuration

There are two implementations:
 - **local client** delivers configuration through go channels directly to the plugins
 - **remote client** stores the configuration using the given `keyval.broker`
