package olog_test

import (
	"testing"

	"ox/pkg/olog"
)

func Test_Info(t *testing.T) {
	olog.Info("hello", olog.Any("a", "b"))
}
