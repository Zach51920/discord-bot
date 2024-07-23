package events

func hasDeveloperRole(roles []string) bool {
	for _, role := range roles {
		if role == "1263315347344457851" || role == "1263639387623915641" {
			return true
		}
	}
	return false
}
