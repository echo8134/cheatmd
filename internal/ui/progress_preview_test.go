package ui

import (
	"strings"
	"testing"

	"github.com/cheatmd-dev/cheatmd/pkg/config"
	"github.com/cheatmd-dev/cheatmd/pkg/parser"
)

func TestProgressPreviewPreservesInsertedReferences(t *testing.T) {
	previousSyntax := config.Get().VarSyntax
	t.Cleanup(func() { config.Get().VarSyntax = previousSyntax })
	for _, tc := range []struct{ name, syntax, command, value, want string }{
		{"dollar", "dollar", `echo "$price" '$unit' $pending`, "$unit 5", `echo "$unit 5" 'USD' $pending`},
		{"angle", "angle", `echo "<price>" '<unit>' <pending>`, "<unit> 5", `echo "<unit> 5" 'USD' <pending>`},
		{"both", "both", `echo "$price" '<unit>' <pending>`, "<unit> $unit 5", `echo "<unit> $unit 5" 'USD' <pending>`},
		{"plain value", "dollar", `echo "$price" '$unit' $pending`, "10", `echo "10" 'USD' $pending`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config.Get().VarSyntax = tc.syntax
			m := &mainModel{varState: &varResolveState{
				cheat:      &parser.Cheat{Command: tc.command},
				currentIdx: 2,
				vars: []varState{
					{def: parser.VarDef{Name: "price"}, resolved: true, value: tc.value},
					{def: parser.VarDef{Name: "unit"}, resolved: true, value: "USD"},
					{def: parser.VarDef{Name: "pending"}},
				},
			}}
			header := StripANSI(m.renderVarHeader(80))
			line, _, _ := strings.Cut(header, "\n")
			if line != tc.want {
				t.Fatalf("progress command = %q, want %q", line, tc.want)
			}
		})
	}
}
