package olog_test

import (
	"testing"

	"github.com/xqk/ox/pkg/olog"
)

func Test_Info(t *testing.T) {
	olog.Info("hello", olog.Any("a", "b"))
}
