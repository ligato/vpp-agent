## pkgreflect - A Go preprocessor for package scoped reflection

Problem: Go reflection does not support enumerating types, variables and functions of packages.

pkgreflect generates a file named pkgreflect.go in every parsed package directory.
This file contains the following maps of exported names to reflection types/values:

	var Types = map[string]reflect.Type{ ... }
	var Functions = map[string]reflect.Value{ ... }
	var Variables = map[string]reflect.Value{ ... }
	var Consts = map[string]reflect.Value{ ... }

Command line usage:

	pkgreflect --help
	pkgreflect [-notypes][-nofuncs][-novars][-noconsts][-unexported][-norecurs][-gofile=filename.go] [DIR_NAME]

If -norecurs is not set, then pkgreflect traverses recursively into sub-directories.
If no DIR_NAME is given, then the current directory is used as root.
