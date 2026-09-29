package executor

import (
	"testing"

	"github.com/cheatmd-dev/cheatmd/pkg/parser"
)

func TestSubstitutionPreservesValuesAndReferences(t *testing.T) {
	for _, tc := range []struct {
		name, input, syntax, want string
		scope                     map[string]string
	}{
		{"identifier", "$user $username", "dollar", "bob $username", map[string]string{"user": "bob"}},
		{"inserted dollar", "$price $unit", "dollar", "$unit 5 USD", map[string]string{"price": "$unit 5", "unit": "USD"}},
		{"inserted angle", "<price> <unit>", "angle", "<unit> 5 USD", map[string]string{"price": "<unit> 5", "unit": "USD"}},
		{"escaped", `echo $user \$user`, "dollar", `echo bob \$user`, map[string]string{"user": "bob"}},
		{"even backslashes", `echo \\$user`, "dollar", `echo \\bob`, map[string]string{"user": "bob"}},
		{"braces", "${user} $user", "both", "${user} bob", map[string]string{"user": "bob"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SubstituteVars(tc.input, tc.scope, tc.syntax); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFinalCommandUnescapesOnlyTemplate(t *testing.T) {
	setVarSyntax(t, "dollar")
	c := &parser.Cheat{Command: `echo $value \$HOME`, Scope: map[string]string{"value": `\$HOME`}}
	if got := NewExecutor(parser.NewCheatIndex()).BuildFinalCommand(c); got != `echo \$HOME $HOME` {
		t.Fatalf("changed inserted value: %q", got)
	}
}

func TestReplaceVarPreservesReferences(t *testing.T) {
	for _, tc := range []struct{ name, input, syntax, want string }{
		{"complete identifier", "$user $username", "dollar", "bob $username"},
		{"escaped dollar", `echo $user \$user`, "dollar", `echo bob \$user`},
		{"both syntaxes", "$user <user> <username>", "both", "bob bob <username>"},
		{"angle only", "$user <user>", "angle", "$user bob"},
		{"dollar only", "$user <user>", "dollar", "bob <user>"},
		{"default syntax", "$user <user>", "", "bob <user>"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ReplaceVar(tc.input, "user", "bob", tc.syntax); got != tc.want {
				t.Fatalf("ReplaceVar() = %q, want %q", got, tc.want)
			}
		})
	}
}
