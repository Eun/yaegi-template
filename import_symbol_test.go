package yaegi_template

import "testing"

func Test_importSymbol_Equals(t *testing.T) {
	tests := []struct {
		name string
		src  importSymbol
		args importSymbol
		want bool
	}{
		{
			"normal import",
			importSymbol{
				Name: "",
				Path: "fmt",
			},
			importSymbol{
				Name: "",
				Path: "fmt",
			},
			true,
		},
		{
			"normal import",
			importSymbol{
				Name: "",
				Path: "fmt",
			},
			importSymbol{
				Name: "",
				Path: "log",
			},
			false,
		},
		{
			"dot import",
			importSymbol{
				Name: ".",
				Path: "fmt",
			},
			importSymbol{
				Name: ".",
				Path: "fmt",
			},
			true,
		},
		{
			"dot import",
			importSymbol{
				Name: "",
				Path: "fmt",
			},
			importSymbol{
				Name: ".",
				Path: "log",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.src.Equals(tt.args); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_importSymbol_ImportLine(t *testing.T) {
	tests := []struct {
		name string
		src  importSymbol
		want string
	}{
		{
			"normal import",
			importSymbol{
				Name: "",
				Path: "fmt",
			},
			`"fmt"`,
		},
		{
			"dot import",
			importSymbol{
				Name: ".",
				Path: "fmt",
			},
			`. "fmt"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.src.ImportLine(); got != tt.want {
				t.Errorf("ImportLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_importSymbols_Contains(t *testing.T) {
	tests := []struct {
		name string
		is   importSymbols
		args importSymbol
		want bool
	}{
		{
			"normal import",
			importSymbols{
				{
					Name: "",
					Path: "fmt",
				},
			},
			importSymbol{
				Name: "",
				Path: "fmt",
			},
			true,
		},
		{
			"normal import",
			importSymbols{
				{
					Name: "",
					Path: "fmt",
				},
			},
			importSymbol{
				Name: "",
				Path: "log",
			},
			false,
		},
		{
			"dot import",
			importSymbols{
				{
					Name: ".",
					Path: "fmt",
				},
			},
			importSymbol{
				Name: ".",
				Path: "fmt",
			},
			true,
		},
		{
			"dot import",
			importSymbols{
				{
					Name: ".",
					Path: "fmt",
				},
			},
			importSymbol{
				Name: ".",
				Path: "log",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.is.Contains(tt.args); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_importSymbols_ImportBlock(t *testing.T) {
	tests := []struct {
		name string
		is   importSymbols
		want string
	}{
		{
			"no imports",
			importSymbols{},
			``,
		},
		{
			"no imports - nil",
			nil,
			``,
		},
		{
			"one normal import",
			importSymbols{
				{
					Name: "",
					Path: "fmt",
				},
			},
			`import "fmt"`,
		},
		{
			"one dot import",
			importSymbols{
				{
					Name: ".",
					Path: "fmt",
				},
			},
			`import . "fmt"`,
		},

		{
			"two normal imports",
			importSymbols{
				{
					Name: "",
					Path: "fmt",
				},
				{
					Name: "",
					Path: "log",
				},
			},
			`import (
"fmt"
"log"
)`,
		},

		{
			"two dot imports",
			importSymbols{
				{
					Name: ".",
					Path: "fmt",
				},
				{
					Name: ".",
					Path: "log",
				},
			},
			`import (
. "fmt"
. "log"
)`,
		},

		{
			"two mixed imports",
			importSymbols{
				{
					Name: "",
					Path: "fmt",
				},
				{
					Name: ".",
					Path: "log",
				},
			},
			`import (
"fmt"
. "log"
)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.is.ImportBlock(); got != tt.want {
				t.Errorf("ImportBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}
