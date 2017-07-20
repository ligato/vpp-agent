# Cassandra implementation of Broker interface

The API was tested with Cassandra 3 and supports:
- UDT (User Defined Types) / embedded structs and honors gocql.Marshaler/gocql.Unmarshaler
- handling all primitive types (like int aliases , IP address); 
  - net.IP can be stored as ipnet
  - net.IPNet can be stored with a MarshalCQL/UnmarshalCQL wrapper go structure
- dumping all rows except for Table
- quering by secondary indexes
- mocking of gocql behavior (using gockle library)in automated unit tests 
