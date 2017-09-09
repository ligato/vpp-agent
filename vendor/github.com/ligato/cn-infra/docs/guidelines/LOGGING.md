To be able to change log levels individually per specific logger using [rpc](../../logging/logmanager):
1. Each plugin is supposed to use it's own logger (injected as dependency) See `plugin.Log` in [examples/logs_plugin](../../examples/logs_plugin/main.go). 
2. If plugin is more complicated then it need to use multiple loggers. See `childLogger` in [examples/logs_plugin](../../examples/logs_plugin/main.go)
