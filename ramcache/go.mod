module github.com/bartventer/gocache/ramcache

go 1.22

toolchain go1.22.4

replace github.com/bartventer/gocache => ../

require (
	github.com/bartventer/gocache v1.15.0
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.6.0
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract [v1.0.0, v1.15.0] // Public API is not stable yet.
