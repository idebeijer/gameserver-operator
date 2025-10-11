package version

// variables overwritten by -ldflags -X at build time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	BuiltBy = "local"
)

func Short() string {
	return Version
}

func All() map[string]string {
	return map[string]string{
		"version": Version,
		"commit":  Commit,
		"date":    Date,
		"builtBy": BuiltBy,
	}
}
