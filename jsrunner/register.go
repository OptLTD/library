package js

import "jsmodule"

const superuserModulePrefix = "superuser/"

// Register registers a native module importable as "superuser/<name>" (legacy superuser-api convention).
func Register(name string, mod modules.Module) {
	modules.RegisterExact(superuserModulePrefix+name, mod)
}

// Module is the native module interface (alias for modules.Module).
type Module = modules.Module

// Global is a namespace of lazy global modules.
type Global = modules.Global
