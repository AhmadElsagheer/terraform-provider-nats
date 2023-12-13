package nats

func invertMap[A comparable, B comparable](in map[A]B) map[B]A {
	out := make(map[B]A, len(in))
	for k, v := range in {
		out[v] = k
	}
	return out
}

func mapFn[A comparable, B any](m map[A]B) func(A) B {
	return func(a A) B { return m[a] }
}
