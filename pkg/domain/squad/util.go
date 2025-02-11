package squad_common

func Unique(values []string) []string {
	uniqueValues := make([]string, 0)
	seen := make(map[string]struct{})
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		uniqueValues = append(uniqueValues, value)
	}
	return uniqueValues
}
