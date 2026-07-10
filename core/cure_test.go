package core

import (
	"testing"
	"time"

	"github.com/sansecio/yargo/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cnaize/cure/source"
)

func TestCureHappyPath(t *testing.T) {
	t.Parallel()

	cure := NewCure().WithSources(source.NewLocal("../testdata/test-rules.zip"))
	err := cure.Run(t.Context())
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	testData := `{"name": "bob", "age": "Valid_Rule_4"}`

	var matches scanner.MatchRules
	err = cure.Scan([]byte(testData), 0, time.Second, &matches)
	assert.NoError(t, err)
	assert.Len(t, matches, 1)
}

func TestCureUpdate(t *testing.T) {
	tests := []struct {
		name     string
		sources  []source.Source
		numrules int
		err      error
	}{
		{
			name:     "valid local source file",
			sources:  []source.Source{source.NewLocal("../testdata/test-rules-1.yar")},
			numrules: 3,
			err:      nil,
		},
		{
			name:     "valid local source directory",
			sources:  []source.Source{source.NewLocal("../testdata")},
			numrules: 4,
			err:      nil,
		},
		{
			name:     "valid local source archive",
			sources:  []source.Source{source.NewLocal("../testdata/test-rules.zip")},
			numrules: 4,
			err:      nil,
		},
		{
			name:     "valid remote source archive",
			sources:  []source.Source{source.NewRemote("https://github.com/YARAHQ/yara-forge/releases/latest/download/yara-forge-rules-core.zip")},
			numrules: -1,
			err:      nil,
		},
		{
			name:     "invalid local source path",
			sources:  []source.Source{source.NewLocal("absent.yara")},
			numrules: 0,
			err:      nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cure := NewCure().WithSources(test.sources...)
			err := cure.Update(t.Context())
			if test.err != nil {
				assert.ErrorIs(t, err, test.err)
			} else {
				assert.NoError(t, err)
			}

			if test.numrules < 0 {
				assert.Greater(t, cure.rules.Load().NumRules(), 1)
			} else {
				assert.Equal(t, test.numrules, cure.rules.Load().NumRules())
			}
		})
	}
}
