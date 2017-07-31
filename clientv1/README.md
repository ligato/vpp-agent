# Client v1

Client v1 (i.e. the first version) defines API that allows to manage configuration of default plugins and linux plugin.
The API can be split into two groups:
 - **resync** applies the given configuration. If a configuration exists it is replaced. (The name is abbreviation
  of *resynchronization* that is required if the previous config is not known.) 
 - **data change** allows to deliver incremental changes of a configuration
