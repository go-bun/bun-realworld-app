package testbed

import (
	"github.com/onsi/gomega/gstruct"
)

func ExtendKeys(a, b gstruct.Keys) gstruct.Keys {
	res := make(gstruct.Keys)
	for k, v := range a {
		res[k] = v
	}
	for k, v := range b {
		res[k] = v
	}
	return res
}
