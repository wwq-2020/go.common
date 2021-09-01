package env

// IsValid IsValid
func IsValid(env Env) bool {
	_, exist := Env_name[int32(env)]
	if !exist {
		return false
	}
	return true
}
