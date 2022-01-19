package testdata

//go:generate protoc --proto_path=. --proto_path=../../../proto --go_out=paths=source_relative:. proto/simple.proto
//go:generate protoc --proto_path=. --proto_path=../../../proto --proto_path=/usr/local/include --go_out=paths=source_relative:. proto/withoption.proto
