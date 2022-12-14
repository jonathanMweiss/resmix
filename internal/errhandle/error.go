package errhandle

import (
	"errors"
	"strings"
)

func AppendError(errs ...error) error {
	sbuilder := strings.Builder{}
	for i, err := range errs {
		if err == nil {
			continue
		}
		sbuilder.WriteString(err.Error())
		if i != len(errs)-1 {
			sbuilder.WriteRune(' ')
		}
	}
	return errors.New(sbuilder.String())
}
