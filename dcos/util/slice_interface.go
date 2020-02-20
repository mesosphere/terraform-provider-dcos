package util

// InterfaceSliceInt32 transforms a []interface{} into []int32
func InterfaceSliceInt32(s []interface{}) (r []int32, ok bool) {
	ok = false
	r = make([]int32, 0)
	for _, v := range s {
		if val, k := v.(int); k {
			// we found at least one value
			ok = true
			r = append(r, int32(val))
		}
	}

	return r, ok
}

// InterfaceSliceString transforms a []interface{} into []string
func InterfaceSliceString(s []interface{}) (r []string, ok bool) {
	ok = false
	r = make([]string, 0)
	for _, v := range s {
		if val, k := v.(string); k {
			// we found at least one value
			ok = true
			r = append(r, val)
		}
	}

	return r, ok
}
