package converter

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_Convert(t *testing.T) {
	// TODO: add more special cases
	tests := []struct {
		cmdName string
	}{
		{cmdName: "1_long"},       // only long
		{cmdName: "2_short_long"}, // short or/and long option
		{cmdName: "3_old"},        // old style option (-opt)
		{cmdName: "4_shorts"},     // multiple short option
	}
	for _, tt := range tests {
		t.Run(tt.cmdName, func(t *testing.T) {
			srcFish, err := os.Open(fmt.Sprintf("./testdata/compfile/%s/%s.fish", tt.cmdName, tt.cmdName))
			require.NoError(t, err)

			dstZsh, err := os.Open(fmt.Sprintf("./testdata/compfile/%s/_%s", tt.cmdName, tt.cmdName))
			require.NoError(t, err)
			wantZsh, err := io.ReadAll(dstZsh)
			require.NoError(t, err)

			c := NewConverter(srcFish, tt.cmdName)
			got, err := c.Convert()
			require.NoError(t, err)

			if strings.TrimSpace(got) != strings.TrimSpace(string(wantZsh)) {
				t.Errorf("got:\n%s\n\n\nwant:\n%s\n\n\ndiff: %s", got, string(wantZsh), cmp.Diff(strings.TrimSpace(got), strings.TrimSpace(string(wantZsh))))
			}
		})
	}
}

func Test_escapeDescMsg(t *testing.T) {
	tests := []struct {
		name    string
		optText string
		want    string
	}{
		{
			"'single_quotes",
			`don\'t`,
			`don'"'"'t`,
		},
		{
			"[]",
			`[MacOS only] hello`,
			`\[MacOS only\] hello`,
		},
		{
			"trim_space",
			`   Lorem  ipsum dolor sit amet,   consectetur adipiscing elit `,
			`Lorem ipsum dolor sit amet, consectetur adipiscing elit`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, escapeDescMsg(tt.optText))
		})
	}
}
