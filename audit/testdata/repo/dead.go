package repo

// dead references linters outside every valid set; each directive must be
// reported with its file and line.
func dead() {
	_ = 1 //nolint:exhaustruct // renamed upstream
	_ = 2 //nolint:exhaustruct,gosec // mixed list: one dead, one live
}
