package util

// Yaml is a struct representation of a yaml, persisted or not
type Yaml struct {
	Content  string
	Filename string
}

// ImageInfo contains identifying information about the target image for a tuber apply
type ImageInfo struct {
	Name string
	Tag string
}
