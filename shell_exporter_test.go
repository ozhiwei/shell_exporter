package main

import (
	"testing"
)

var config = &Config{
	Shells: []*Shell{},
}

func TestShellManager(t *testing.T) {
	t.Run("Check ShellManger", func(t *testing.T) {
		shellManager, err := NewShellManger()
		if err != nil {
			t.Errorf("NewShellManger Fail")
		} else {
			shellManager.initShellManger()
		}
	})
}

func TestShell(t *testing.T) {
	t.Run("Check Shell", func(t *testing.T) {
		shell := Shell{
			Name: "shell_exporter_test",
			Help: "shell exporter test metric",
			Cmd: "hostname; echo $RANDOM",
			ConstLabels: map[string]string{
				"env": "test",
				"app": "shell_exporter",
			},
			LabelsRegexp: "(?P<hostname>.+)\n(?P<value>[0-9.]+)",
			Bin: "/bin/bash",
		}

		shell.init()
		shell.run()
		shell.match()
		shell.collect()
		t.Log("shell metric: ", shell)
	})
}
