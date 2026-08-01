package skills

// Clone returns a deep-cloned copy of the provided skills slice.
func Clone(items []Skill) []Skill {
	return cloneSkills(items)
}

// CloneLoadReport returns a detached copy of a load report.
func CloneLoadReport(report LoadReport) LoadReport {
	return cloneLoadReport(report)
}

// LoadGeneration returns the current loader generation for the provided options.
func LoadGeneration(opts LoadOptions) (uint64, bool) {
	sources := buildLoadSources(opts)
	key, ok := buildLoadCacheLookupKey(sources, opts)
	if !ok {
		return 0, false
	}
	return resolveLoadGeneration(key, sources, opts)
}
