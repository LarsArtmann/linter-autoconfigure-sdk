package repo

// ok carries only nolint directives referencing valid names: enabled,
// default, and the special name all.
func ok() {
	_ = 1 //nolint:gosec // enabled in valid.golangci.yml
	_ = 2 //nolint:staticcheck // enabled by default
	_ = 3 //nolint:all // the special name
	_ = 4 // nolint:gosec,all // spaced and comma-separated
}
