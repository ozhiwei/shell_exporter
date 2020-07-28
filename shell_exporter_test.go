package main

import (
	"testing"
)

var config = &Config{
	Shells: []*Shell{},
}

func TestShellManager(t *testing.T) {
	t.Run("RequiredParameters", func(t *testing.T) {
		t.Log("hello world!")
	})
}
