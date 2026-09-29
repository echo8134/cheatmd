package main

import (
	"testing"

	"github.com/cheatmd-dev/cheatmd/pkg/config"
)

func TestComposeRewritesCompleteReferences(t *testing.T) {
	old := config.Get().VarSyntax
	t.Cleanup(func() { config.Get().VarSyntax = old })
	for _, tc := range []struct {
		name, syntax, command, want, definitions string
	}{
		{"overlapping names", "angle", "echo $p $port", "echo <p> <port>", "var p\nvar port\n"},
		{"escaped reference", "angle", `echo \$value $value`, `echo \$value <value>`, "var value\n"},
		{"shell expansion", "angle", "echo ${value:-fallback} $value", "echo ${value:-fallback} <value>", "var value\n"},
		{"dollar default", "dollar", "echo <host> <port>", "echo $host $port", "var host\nvar port\n"},
		{"empty syntax", "", "echo <host>", "echo $host", "var host\n"},
		{"both preserved", "both", "echo $host <port>", "echo $host <port>", "var host\nvar port\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config.Get().VarSyntax = tc.syntax
			got := buildComposeMarkdown("Example", "", tc.command)
			want := "\n# Example\n\n```sh title:\"\"\n" + tc.want + "\n```\n<!-- cheat\n" + tc.definitions + "-->\n"
			if got != want {
				t.Fatalf("composed Markdown = %q, want %q", got, want)
			}
		})
	}
}
