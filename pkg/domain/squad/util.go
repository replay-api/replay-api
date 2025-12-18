package squad

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

func RolesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, role := range a {
		found := false
		for _, otherRole := range b {
			if role == otherRole {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
