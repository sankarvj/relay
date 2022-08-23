package forms

func keyMap(namedKeys map[string]string, namedVals map[string]interface{}) map[string]interface{} {
	itemVals := make(map[string]interface{}, 0)
	for name, key := range namedKeys {
		itemVals[key] = namedVals[name]
	}
	return itemVals
}
