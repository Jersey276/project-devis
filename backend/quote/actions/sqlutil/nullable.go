package sqlutil

func NullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
