package sqlutil

import "github.com/lib/pq"

// StringArray wraps a []string so it can be passed to a Postgres `= ANY($n)`
// clause via the lib/pq driver.
func StringArray(v []string) interface{} {
	return pq.Array(v)
}

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
