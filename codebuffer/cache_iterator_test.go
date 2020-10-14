package codebuffer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCacheIterator(t *testing.T) {
	tests := []struct {
		Name  string
		Parts []*Part
	}{
		{
			"0 Parts",
			nil,
		},
		{
			"1 Part",
			[]*Part{
				{
					Type:    CodePartType,
					Content: []byte("Foo"),
				},
			},
		},
		{
			"2 Parts",
			[]*Part{
				{
					Type:    CodePartType,
					Content: []byte("Foo"),
				},
				{
					Type:    CodePartType,
					Content: []byte("Bar"),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			it, err := newCacheIterator(test.Parts)
			require.NoError(t, err)
			require.Nil(t, it.Value())
			require.NoError(t, it.Error())

			for i := 0; i < len(test.Parts); i++ {
				require.True(t, it.Next(), i)
				require.Equal(t, test.Parts[i], it.Value(), i)
			}

			require.False(t, it.Next())
			require.Nil(t, it.Value())
			require.NoError(t, it.Error())
		})
	}
}
