package utils

func SlicesFilterFunc[S ~[]E, E any](s S, fnShouldKeep func(E) bool) S {
	o := S{}

	for _, v := range s {
		if fnShouldKeep(v) {
			o = append(o, v)
		}
	}

	return o
}
