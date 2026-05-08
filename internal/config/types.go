package config

type GlobalConfig struct {
	Defaults DefaultsConfig
	JDKs     map[string]string
	Mavens   map[string]string
}

type DefaultsConfig struct {
	JDK       string `toml:"jdk"`
	MavenHome string `toml:"maven_home"`
	Settings  string `toml:"settings"`
	LocalRepo string `toml:"local_repo"`
}

type ProjectConfig struct {
	JDK       string `toml:"jdk"`
	Maven     string `toml:"maven"`
	Settings  string `toml:"settings"`
	LocalRepo string `toml:"local_repo"`
}

type ResolvedConfig struct {
	JavaCmd    string
	MavenHome  string
	Settings   string
	LocalRepo  string
	ProjectDir string

	JavaCmdSource   string
	MavenHomeSource string
	SettingsSource  string
	LocalRepoSource string
}
