package env

// IsValid IsValid
func IsValid(env Env) bool {
	_, exist := Env_name[int32(env)]
	return exist
}
