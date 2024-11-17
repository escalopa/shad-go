//go:build !solution

package varjoin

import "strings"

func Join(sep string, args ...string) string {
	if len(args) == 0 {
		return ""
	}

	bytes := len(sep) * (len(args) - 1)

	for _, arg := range args {
		bytes += len(arg)
	}

	var b strings.Builder
	b.Grow(bytes)

	for i, arg := range args {
		b.WriteString(arg)
		if i < len(args)-1 {
			b.WriteString(sep)
		}
	}

	return b.String()
}
