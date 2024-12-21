package testutil

import (
	"strings"
	"testing"

	"github.com/vladimirvivien/automi/api"
)

func LogSinkFunc(t *testing.T) func(api.StreamLog) error {
	return func(log api.StreamLog) error {
		var builder strings.Builder
		builder.WriteString(log.Message)
		for _, attr := range log.Attrs {
			builder.WriteString(" ")
			builder.WriteString(attr.Key)
			builder.WriteString("=")
			builder.WriteString(attr.Value.String())
		}
		t.Log(builder.String())
		return nil
	}
}
