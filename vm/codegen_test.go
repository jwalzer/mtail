// Copyright 2011 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package vm

import (
	"strings"
	"testing"
	"time"

	go_cmp "github.com/google/go-cmp/cmp"
)

var testCodeGenPrograms = []struct {
	name   string
	source string
	prog   []instr // expected bytecode
}{
	// Composite literals require too many explicit conversions.
	{"simple line counter",
		"counter line_count\n/$/ { line_count++\n }\n",
		[]instr{
			{match, 0},
			{jnm, 7},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true}}},
	{"count a",
		"counter a_count\n/a$/ { a_count++\n }\n",
		[]instr{
			{match, 0},
			{jnm, 7},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true}}},
	{"strptime and capref",
		"counter foo\n" +
			"/(.*)/ { strptime($1, \"2006-01-02T15:04:05\")\n" +
			"foo++\n}\n",
		[]instr{
			{match, 0},
			{jnm, 11},
			{setmatched, false},
			{push, 0},
			{capref, 1},
			{str, 0},
			{strptime, 2},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true}}},
	{"strptime and named capref",
		"counter foo\n" +
			"/(?P<date>.*)/ { strptime($date, \"2006-01-02T15:04:05\")\n" +
			"foo++\n }\n",
		[]instr{
			{match, 0},
			{jnm, 11},
			{setmatched, false},
			{push, 0},
			{capref, 1},
			{str, 0},
			{strptime, 2},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true}}},
	{"inc by and set",
		"counter foo\ncounter bar\n" +
			"/([0-9]+)/ {\n" +
			"foo += $1\n" +
			"bar = $1\n" +
			"}\n",
		[]instr{
			{match, 0},
			{jnm, 16},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{push, 0},
			{capref, 1},
			{s2i, nil},
			{inc, 0},
			{mload, 1},
			{dload, 0},
			{push, 0},
			{capref, 1},
			{s2i, nil},
			{iset, nil},
			{setmatched, true}}},
	{"cond expr gt",
		"counter foo\n" +
			"1 > 0 {\n" +
			"  foo++\n" +
			"}\n",
		[]instr{
			{push, int64(1)},
			{push, int64(0)},
			{icmp, 1},
			{jnm, 6},
			{push, true},
			{jmp, 7},
			{push, false},
			{jnm, 13},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true}}},
	{"cond expr lt",
		"counter foo\n" +
			"1 < 0 {\n" +
			"  foo++\n" +
			"}\n",
		[]instr{
			{push, int64(1)},
			{push, int64(0)},
			{icmp, -1},
			{jnm, 6},
			{push, true},
			{jmp, 7},
			{push, false},
			{jnm, 13},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true}}},
	{"cond expr eq",
		"counter foo\n" +
			"1 == 0 {\n" +
			"  foo++\n" +
			"}\n",
		[]instr{
			{push, int64(1)},
			{push, int64(0)},
			{icmp, 0},
			{jnm, 6},
			{push, true},
			{jmp, 7},
			{push, false},
			{jnm, 13},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true}}},
	{"cond expr le",
		"counter foo\n" +
			"1 <= 0 {\n" +
			"  foo++\n" +
			"}\n",
		[]instr{
			{push, int64(1)},
			{push, int64(0)},
			{icmp, 1},
			{jm, 6},
			{push, true},
			{jmp, 7},
			{push, false},
			{jnm, 13},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true}}},
	{"cond expr ge",
		"counter foo\n" +
			"1 >= 0 {\n" +
			"  foo++\n" +
			"}\n",
		[]instr{
			{push, int64(1)},
			{push, int64(0)},
			{icmp, -1},
			{jm, 6},
			{push, true},
			{jmp, 7},
			{push, false},
			{jnm, 13},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true}}},
	{"cond expr ne",
		"counter foo\n" +
			"1 != 0 {\n" +
			"  foo++\n" +
			"}\n",
		[]instr{
			{push, int64(1)},
			{push, int64(0)},
			{icmp, 0},
			{jm, 6},
			{push, true},
			{jmp, 7},
			{push, false},
			{jnm, 13},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true}}},
	{"nested cond",
		"counter foo\n" +
			"/(\\d+)/ {\n" +
			"  $1 <= 1 {\n" +
			"    foo++\n" +
			"  }\n" +
			"}\n",
		[]instr{
			{match, 0},
			{jnm, 19},
			{setmatched, false},
			{push, 0},
			{capref, 1},
			{s2i, nil},
			{push, int64(1)},
			{icmp, 1},
			{jm, 11},
			{push, true},
			{jmp, 12},
			{push, false},
			{jnm, 18},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true},
			{setmatched, true}}},
	{"deco",
		"counter foo\n" +
			"counter bar\n" +
			"def fooWrap {\n" +
			"  /.*/ {\n" +
			"    foo++\n" +
			"    next\n" +
			"  }\n" +
			"}\n" +
			"" +
			"@fooWrap { bar++\n }\n",
		[]instr{
			{match, 0},
			{jnm, 10},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{mload, 1},
			{dload, 0},
			{inc, nil},
			{setmatched, true}}},
	{"length",
		"len(\"foo\") > 0 {\n" +
			"}\n",
		[]instr{
			{str, 0},
			{length, 1},
			{push, int64(0)},
			{cmp, 1},
			{jnm, 7},
			{push, true},
			{jmp, 8},
			{push, false},
			{jnm, 11},
			{setmatched, false},
			{setmatched, true}}},
	{"bitwise", `
1 & 7 ^ 15 | 8
~ 16 << 2
1 >> 20
`,
		[]instr{
			{push, int64(1)},
			{push, int64(7)},
			{and, nil},
			{push, int64(15)},
			{xor, nil},
			{push, int64(8)},
			{or, nil},
			{push, int64(16)},
			{neg, nil},
			{push, int64(2)},
			{shl, nil},
			{push, int64(1)},
			{push, int64(20)},
			{shr, nil}}},
	{"pow", `
/(\d+) (\d+)/ {
$1 ** $2
}
`,
		[]instr{
			{match, 0},
			{jnm, 11},
			{setmatched, false},
			{push, 0},
			{capref, 1},
			{s2i, nil},
			{push, 0},
			{capref, 2},
			{s2i, nil},
			{ipow, nil},
			{setmatched, true}}},
	{"indexed expr", `
counter a by b
a["string"]++
`,
		[]instr{
			{str, 0},
			{mload, 0},
			{dload, 1},
			{inc, nil}}},
	{"strtol", `
strtol("deadbeef", 16)
`,
		[]instr{
			{str, 0},
			{push, int64(16)},
			{s2i, 2}}},
	{"float", `
20.0
`,
		[]instr{
			{push, 20.0}}},
	{"otherwise", `
counter a
otherwise {
	a++
}
`,
		[]instr{
			{otherwise, nil},
			{jnm, 7},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true}}},
	{"cond else",
		`counter foo
counter bar
1 > 0 {
  foo++
} else {
  bar++
}`,
		[]instr{
			{push, int64(1)},
			{push, int64(0)},
			{icmp, 1},
			{jnm, 6},
			{push, true},
			{jmp, 7},
			{push, false},
			{jnm, 14},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true},
			{jmp, 17},
			{mload, 1},
			{dload, 0},
			{inc, nil},
		},
	},
	{"mod",
		`
3 % 1
`,
		[]instr{
			{push, int64(3)},
			{push, int64(1)},
			{imod, nil},
		},
	},
	{"del", `
counter a by b
del a["string"]
`,
		[]instr{
			{str, 0},
			{mload, 0},
			{del, 1}},
	},
	{"del after", `
counter a by b
del a["string"] after 1h
`,
		[]instr{
			{push, time.Hour},
			{str, 0},
			{mload, 0},
			{expire, 1}},
	},
	{"types", `
gauge i
gauge f
/(\d+)/ {
 i = $1
}
/(\d+\.\d+)/ {
 f = $1
}
`,
		[]instr{
			{match, 0},
			{jnm, 10},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{push, 0},
			{capref, 1},
			{s2i, nil},
			{iset, nil},
			{setmatched, true},
			{match, 1},
			{jnm, 20},
			{setmatched, false},
			{mload, 1},
			{dload, 0},
			{push, 1},
			{capref, 1},
			{s2f, nil},
			{fset, nil},
			{setmatched, true},
		},
	},

	{"getfilename", `
getfilename()
`,
		[]instr{
			{getfilename, 0},
		},
	},

	{"dimensioned counter",
		`counter c by a,b,c
/(\d) (\d) (\d)/ {
  c[$1,$2][$3]++
}
`,
		[]instr{
			{match, 0},
			{jnm, 19},
			{setmatched, false},
			{push, 0},
			{capref, 1},
			{s2i, nil},
			{i2s, nil},
			{push, 0},
			{capref, 2},
			{s2i, nil},
			{i2s, nil},
			{push, 0},
			{capref, 3},
			{s2i, nil},
			{i2s, nil},
			{mload, 0},
			{dload, 3},
			{inc, nil},
			{setmatched, true}}},
	{"string to int",
		`counter c
/(.*)/ {
  c = int($1)
}
`,
		[]instr{
			{match, 0},
			{jnm, 10},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{push, 0},
			{capref, 1},
			{s2i, nil},
			{iset, nil},
			{setmatched, true}}},
	{"int to float",
		`counter c
/(\d)/ {
  c = float($1)
}
`,
		[]instr{
			{match, 0},
			{jnm, 11},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{push, 0},
			{capref, 1},
			{s2i, nil},
			{i2f, nil},
			{fset, nil},
			{setmatched, true}}},
	{"string to float",
		`counter c
/(.*)/ {
  c = float($1)
}
`,
		[]instr{
			{match, 0},
			{jnm, 10},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{push, 0},
			{capref, 1},
			{s2f, nil},
			{fset, nil},
			{setmatched, true}}},
	{"float to string",
		`counter c by a
/(\d+\.\d+)/ {
  c[string($1)] ++
}
`,
		[]instr{
			{match, 0},
			{jnm, 11},
			{setmatched, false},
			{push, 0},
			{capref, 1},
			{s2f, nil},
			{f2s, nil},
			{mload, 0},
			{dload, 1},
			{inc, nil},
			{setmatched, true}}},
	{"int to string",
		`counter c by a
/(\d+)/ {
  c[string($1)] ++
}
`,
		[]instr{
			{match, 0},
			{jnm, 11},
			{setmatched, false},
			{push, 0},
			{capref, 1},
			{s2i, nil},
			{i2s, nil},
			{mload, 0},
			{dload, 1},
			{inc, nil},
			{setmatched, true}}},
	{"nested comparisons",
		`counter foo
/(.*)/ {
  $1 == "foo" || $1 == "bar" {
    foo++
  }
}
`, []instr{
			{match, 0},
			{jnm, 31},
			{setmatched, false},
			{push, 0},
			{capref, 1},
			{str, 0},
			{scmp, 0},
			{jnm, 10},
			{push, true},
			{jmp, 11},
			{push, false},
			{jm, 23},
			{push, 0},
			{capref, 1},
			{str, 1},
			{scmp, 0},
			{jnm, 19},
			{push, true},
			{jmp, 20},
			{push, false},
			{jm, 23},
			{push, false},
			{jmp, 24},
			{push, true},
			{jnm, 30},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true},
			{setmatched, true}}},
	{"string concat", `
counter f by s
/(.*), (.*)/ {
  f[$1 + $2]++
}
`,
		[]instr{
			{match, 0},
			{jnm, 12},
			{setmatched, false},
			{push, 0},
			{capref, 1},
			{push, 0},
			{capref, 2},
			{cat, nil},
			{mload, 0},
			{dload, 1},
			{inc, nil},
			{setmatched, true},
		}},
	{"add assign float", `
gauge foo
/(\d+\.\d+)/ {
  foo += $1
}
`,
		[]instr{
			{match, 0},
			{jnm, 13},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{mload, 0},
			{dload, 0},
			{push, 0},
			{capref, 1},
			{s2f, nil},
			{fadd, nil},
			{fset, nil},
			{setmatched, true},
		}},
	{"match expression", `
	counter foo
	/(.*)/ {
	  $1 =~ /asdf/ {
	    foo++
	  }
	}`,
		[]instr{
			{match, 0},
			{jnm, 13},
			{setmatched, false},
			{push, 0},
			{capref, 1},
			{smatch, 1},
			{jnm, 12},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true},
			{setmatched, true},
		}},
	{"negative match expression", `
	counter foo
	/(.*)/ {
	  $1 !~ /asdf/ {
	    foo++
	  }
	}`,
		[]instr{
			{match, 0},
			{jnm, 14},
			{setmatched, false},
			{push, 0},
			{capref, 1},
			{smatch, 1},
			{not, nil},
			{jnm, 13},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true},
			{setmatched, true},
		}},
	{"capref used in def", `
/(?P<x>\d+)/ && $x > 5 {
}`,
		[]instr{
			{match, 0},
			{jnm, 14},
			{push, 0},
			{capref, 1},
			{s2i, nil},
			{push, int64(5)},
			{icmp, 1},
			{jnm, 10},
			{push, true},
			{jmp, 11},
			{push, false},
			{jnm, 14},
			{push, true},
			{jmp, 15},
			{push, false},
			{jnm, 18},
			{setmatched, false},
			{setmatched, true},
		}},
	{"binop arith type conversion", `
gauge var
/(?P<x>\d+) (\d+\.\d+)/ {
  var = $x + $2
}`,
		[]instr{
			{match, 0},
			{jnm, 15},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{push, 0},
			{capref, 1},
			{s2i, nil},
			{i2f, nil},
			{push, 0},
			{capref, 2},
			{s2f, nil},
			{fadd, nil},
			{fset, nil},
			{setmatched, true},
		}},
	{"binop compare type conversion", `
counter var
/(?P<x>\d+) (\d+\.\d+)/ {
  $x > $2 {
    var++
  }
}`,
		[]instr{
			{match, 0},
			{jnm, 22},
			{setmatched, false},
			{push, 0},
			{capref, 1},
			{s2i, nil},
			{i2f, nil},
			{push, 0},
			{capref, 2},
			{s2f, nil},
			{fcmp, 1},
			{jnm, 14},
			{push, true},
			{jmp, 15},
			{push, false},
			{jnm, 21},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{inc, nil},
			{setmatched, true},
			{setmatched, true},
		}},
	{"set string", `
text foo
/(.*)/ {
  foo = $1
}
`, []instr{
		{match, 0},
		{jnm, 9},
		{setmatched, false},
		{mload, 0},
		{dload, 0},
		{push, 0},
		{capref, 1},
		{sset, nil},
		{setmatched, true},
	}},
	{"concat to text", `
text foo
/(?P<v>.*)/ {
		foo += $v
}`,
		[]instr{
			{match, 0},
			{jnm, 12},
			{setmatched, false},
			{mload, 0},
			{dload, 0},
			{mload, 0},
			{dload, 0},
			{push, 0},
			{capref, 1},
			{cat, nil},
			{sset, nil},
			{setmatched, true},
		}},
	{"decrement", `
counter i
// {
  i--
}`, []instr{
		{match, 0},
		{jnm, 7},
		{setmatched, false},
		{mload, 0},
		{dload, 0},
		{dec, nil},
		{setmatched, true},
	}},
	{"capref and settime", `
/(\d+)/ {
  settime($1)
}`, []instr{
		{match, 0},
		{jnm, 8},
		{setmatched, false},
		{push, 0},
		{capref, 1},
		{s2i, nil},
		{settime, 1},
		{setmatched, true},
	}},
	{"cast to self", `
/(\d+)/ {
settime(int($1))
}`, []instr{
		{match, 0},
		{jnm, 8},
		{setmatched, false},
		{push, 0},
		{capref, 1},
		{s2i, nil},
		{settime, 1},
		{setmatched, true},
	}},
}

func TestCodegen(t *testing.T) {
	for _, tc := range testCodeGenPrograms {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ast, err := Parse(tc.name, strings.NewReader(tc.source))
			if err != nil {
				t.Fatalf("Parse error: %s", err)
			}
			err = Check(ast)
			s := Sexp{}
			s.emitTypes = true
			t.Log("Typed AST:\n" + s.Dump(ast))
			if err != nil {
				t.Fatalf("Check error: %s", err)
			}
			obj, err := CodeGen(tc.name, ast)
			if err != nil {
				t.Fatalf("Codegen error:\n%s", err)
			}

			if diff := go_cmp.Diff(tc.prog, obj.prog, go_cmp.AllowUnexported(instr{})); diff != "" {
				t.Error(diff)
				t.Logf("Expected:\n%s\nReceived:\n%s", tc.prog, obj.prog)
			}
		})
	}
}
