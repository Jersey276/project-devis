package sqlutil

func NullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func NullableInt32(v int32) interface{} {
	if v == 0 {
		return nil
	}
	return v
}
