package vendor

// pruned must never be reported: the vendor directory is skipped.
func pruned() {
	_ = 1 //nolint:exhaustruct
}
