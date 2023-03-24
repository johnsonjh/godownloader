module github.com/goreleaser/godownloader

go 1.13

require (
	github.com/apex/log v1.9.0
	github.com/client9/codegen v0.0.0-20180316044450-92480ce66a06
	github.com/goreleaser/goreleaser v0.184.0
	github.com/pkg/errors v0.9.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.4.0
)

// TODO: remove this when https://github.com/google/rpmpack/pull/33 gets merged in.
replace github.com/google/rpmpack => github.com/caarlos0/rpmpack v0.0.0-20191106130752-24a815bfaee0
