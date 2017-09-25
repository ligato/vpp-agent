# Logs-lib HTTP Example

To run the example, simply type:
```
go run server.go
```

List all registered loggers and their current log level via HTTP GET:
```
curl localhost:8080/list
```

Modify log level remotely via HTTP POST:
```
curl localhost:8080/set/{loggerName}/{logLevel}
```