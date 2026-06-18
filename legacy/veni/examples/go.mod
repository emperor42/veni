module example

go 1.21

// Replace directive points to the local VENI package for development.
// When publishing to production, remove this line and use the actual module path.
replace github.com/Emperor42/veni => ../

require github.com/Emperor42/veni v0.0.0