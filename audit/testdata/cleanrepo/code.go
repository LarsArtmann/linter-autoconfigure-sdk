package cleanrepo

// ok carries only nolint directives referencing names valid under
// valid.golangci.yml.
func ok() {
	_ = 1 //nolint:gosec // enabled
	_ = 2 //nolint:staticcheck // default
	_ = 3 //nolint:all // special name
}
