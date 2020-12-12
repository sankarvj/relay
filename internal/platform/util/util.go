package util

func ConvertSliceType(s []string) []interface{} {
	inf := make([]interface{}, len(s))
	for i, v := range s {
		inf[i] = v
	}
	return inf
}

func ConvertSliceTypeRev(inf []interface{}) []string {
	s := make([]string, len(inf))
	for i, v := range inf {
		s[i] = v.(string)
	}
	return s
}
