package internal

import (
	"os"
	"strings"
)

func InTest() bool {
	for _, arg := range os.Args {
		for _, s := range []string{"test.v", "test.run", "/_test/", ".test"} {
			if strings.Contains(arg, s) {
				return true
			}
		}
	}
	return false
}
