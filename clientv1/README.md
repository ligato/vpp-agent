# Client v1

Client v1 (i.e. the first version) defines an API that allows to manage configuration of default plugins and the Linux plugin.
The way of configuration transport from an API user to the plugins is abstracted.
The API calls can be split into two groups:
 - **resync** applies the given configuration. If a configuration already exists, it is replaced. (The name is an abbreviation
  of *resynchronization* that is required if the previous config is not known.) 
 - **data change** allows to deliver incremental changes of a configuration

There are two implementations:
 - **local client** delivers configuration through go channels directly to the plugins
 - **remote client** stores the configuration using the given `keyval.broker`
