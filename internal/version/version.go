package version

// These variables are set at build time via ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
}

func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
}
