package manifest

type Manifest struct {
	Info Info `yaml:"info"`
}

type Info struct {
	Name string `yaml:"name"`
	Version string `yaml:"version"`
}
