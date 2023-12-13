package provider

import "github.com/hashicorp/terraform-plugin-framework-validators/int64validator"

var infinityOrPositiveInt64Validator = int64validator.Any(int64validator.OneOf(-1), int64validator.AtLeast(1))

func convertSlice[S any, D any](in []S, fn func(S) D) []D {
	out := make([]D, 0, len(in))
	for _, elem := range in {
		out = append(out, fn(elem))
	}
	return out
}
