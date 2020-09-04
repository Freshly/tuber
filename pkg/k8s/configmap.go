package k8s

// PatchConfigMap gets, patches, and saves a configmap
func PatchConfigMap(mapName string, namespace string, key string, value string) (err error) {
	config, err := GetConfigResource(mapName, namespace, "ConfigMap")

	if err != nil {
		return
	}

	config.Data[key] = value

	return config.Save(namespace)
}

// RemoveConfigMapEntry removes an entry, from a configmap
func RemoveConfigMapEntry(mapName string, namespace string, key string) (err error) {
	config, err := GetConfigResource(mapName, namespace, "ConfigMap")

	if err != nil {
		return
	}

	delete(config.Data, key)

	return config.Save(namespace)
}
