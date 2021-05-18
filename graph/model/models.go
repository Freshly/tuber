package model

func (t TuberApp) Indexes() (map[string]string, map[string]bool, map[string]int) {
	return map[string]string{
			"name":          t.Name,
			"imageTag":      t.ImageTag,
			"sourceAppName": t.SourceAppName,
		}, map[string]bool{
			"reviewApp": t.ReviewApp,
		}, map[string]int{}
}

func (t TuberApp) BucketName() string {
	return "apps"
}

func (t TuberApp) Key() string {
	return t.Name
}
