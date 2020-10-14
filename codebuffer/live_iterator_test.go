package codebuffer

import (
	"testing"

	"bytes"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

func TestLiveIterator(t *testing.T) {
	tests := []struct {
		Name          string
		Input         string
		StartSequence []rune
		EndSequence   []rune
		ExpectedParts []*Part
	}{
		{
			"0 Parts",
			"",
			[]rune("<$"),
			[]rune("$>"),
			nil,
		},
		{
			"1 Part",
			"Foo Bar",
			[]rune("<$"),
			[]rune("$>"),
			[]*Part{
				{
					Type:    TextPartType,
					Content: []byte("Foo Bar"),
				},
			},
		},
		{
			"2 Parts",
			"Foo <$ Bar $>",
			[]rune("<$"),
			[]rune("$>"),
			[]*Part{
				{
					Type:    TextPartType,
					Content: []byte("Foo "),
				},
				{
					Type:    CodePartType,
					Content: []byte(" Bar "),
				},
			},
		},
		{
			"2 Parts",
			"<$ Foo $> Bar",
			[]rune("<$"),
			[]rune("$>"),
			[]*Part{
				{
					Type:    CodePartType,
					Content: []byte(" Foo "),
				},
				{
					Type:    TextPartType,
					Content: []byte(" Bar"),
				},
			},
		},
		{
			"3 Parts",
			"Foo <$ Bar $> Baz",
			[]rune("<$"),
			[]rune("$>"),
			[]*Part{
				{
					Type:    TextPartType,
					Content: []byte("Foo "),
				},
				{
					Type:    CodePartType,
					Content: []byte(" Bar "),
				},
				{
					Type:    TextPartType,
					Content: []byte(" Baz"),
				},
			},
		},
		{
			"3 Parts - White Space Removal",
			"Foo <$- Bar -$> Baz",
			[]rune("<$"),
			[]rune("$>"),
			[]*Part{
				{
					Type:    TextPartType,
					Content: []byte("Foo"),
				},
				{
					Type:    CodePartType,
					Content: []byte(" Bar "),
				},
				{
					Type:    TextPartType,
					Content: []byte("Baz"),
				},
			},
		},
		{
			"Only Code Part",
			"<$- Bar -$>",
			[]rune("<$"),
			[]rune("$>"),
			[]*Part{
				{
					Type:    CodePartType,
					Content: []byte(" Bar "),
				},
			},
		},
		{
			"Open Sequence but no Close Sequence",
			"<$- Bar ",
			[]rune("<$"),
			[]rune("$>"),
			[]*Part{
				{
					Type:    CodePartType,
					Content: []byte(" Bar "),
				},
			},
		},
		{
			"Open Sequence but no Close Sequence",
			"Foo <$- Bar ",
			[]rune("<$"),
			[]rune("$>"),
			[]*Part{
				{
					Type:    TextPartType,
					Content: []byte("Foo"),
				},
				{
					Type:    CodePartType,
					Content: []byte(" Bar "),
				},
			},
		},
		{
			"No Start Sequence",
			"Foo Bar",
			[]rune(""),
			[]rune("$>"),
			[]*Part{
				{
					Type:    CodePartType,
					Content: []byte("Foo Bar"),
				},
			},
		},
		{
			"No End Sequence",
			"Foo <$- Bar",
			[]rune("<$"),
			[]rune(""),
			[]*Part{
				{
					Type:    TextPartType,
					Content: []byte("Foo"),
				},
				{
					Type:    CodePartType,
					Content: []byte(" Bar"),
				},
			},
		},
		{
			"Interrupted Start Sequence",
			"Foo <-$ Bar <$ Baz $>",
			[]rune("<$"),
			[]rune("$>"),
			[]*Part{
				{
					Type:    TextPartType,
					Content: []byte("Foo <-$ Bar "),
				},
				{
					Type:    CodePartType,
					Content: []byte(" Baz "),
				},
			},
		},
		{
			"Interrupted End Sequence",
			"Foo <$ Bar $-> Baz $> Taz",
			[]rune("<$"),
			[]rune("$>"),
			[]*Part{
				{
					Type:    TextPartType,
					Content: []byte("Foo "),
				},
				{
					Type:    CodePartType,
					Content: []byte(" Bar $-> Baz "),
				},
				{
					Type:    TextPartType,
					Content: []byte(" Taz"),
				},
			},
		},
		{
			"Interrupted Start Sequence with EOF",
			"Foo <",
			[]rune("<$"),
			[]rune("$>"),
			[]*Part{
				{
					Type:    TextPartType,
					Content: []byte("Foo <"),
				},
			},
		},
		{
			"Interrupted End Sequence with EOF",
			"Foo <$ Bar $",
			[]rune("<$"),
			[]rune("$>"),
			[]*Part{
				{
					Type:    TextPartType,
					Content: []byte("Foo "),
				},
				{
					Type:    CodePartType,
					Content: []byte(" Bar $"),
				},
			},
		},
		{
			"Empty Code Part",
			"Foo <$$> Bar",
			[]rune("<$"),
			[]rune("$>"),
			[]*Part{
				{
					Type:    TextPartType,
					Content: []byte("Foo "),
				},
				{
					Type:    TextPartType,
					Content: []byte(" Bar"),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var parts []*Part
			it, err := newLiveIterator(atomic.NewInt32(0), &parts, bytes.NewReader([]byte(test.Input)), test.StartSequence, test.EndSequence)
			require.NoError(t, err)
			require.NoError(t, it.Error())

			for i := 0; i < len(test.ExpectedParts); i++ {
				require.True(t, it.Next(), i)
				require.Equal(t, test.ExpectedParts[i], it.Value(), i)
			}

			require.False(t, it.Next())
			require.Nil(t, it.Value())
			require.NoError(t, it.Error())
			// test if next is also false
			require.False(t, it.Next())
			require.Nil(t, it.Value())
			require.NoError(t, it.Error())
		})
	}
}
