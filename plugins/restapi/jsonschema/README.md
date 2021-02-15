# Protoc-gen-jsonschema
Content of this package and subpackages is modified code of 3rd party library [protoc-gen-jsonschema](https://github.com/chrusty/protoc-gen-jsonschema).
The purpose of the tool is to provide proto to JSON schema conversion in the form of protoc plugin.
The customization for ligato is not touching the conversion functionality, but only removes the protoc 
dependency and enables it to be used as library (internal packages in original repository).


## Changes tracking
The base code for ligato modifications is [here](https://github.com/chrusty/protoc-gen-jsonschema/tree/de75f1b59c4e0f5d5edf7be2a18d1c8e4d81b17a).
Initial commit changes:

- removed all unnecessary parts 
  - project root files except License 
  - CMD part for protoc connection
  - convertor tests (they are dependent on protoc at test runtime)   
- extracted convertor out of internal package to be able to use it
- relaxed some info level logging to debug level logging (proto_package.go, lines 78 and 121)
- removed "oneof" type from enums to provide compatibility with external json example generator (types.go line 137 and converter.go line 116)

Other changes can be tracked by git changes in this package and its subpackages