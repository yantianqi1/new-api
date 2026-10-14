package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAffiliateCommissionRateBasisPoints(t *testing.T) {
	testCases := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{name: "disabled", value: "0", want: 0},
		{name: "decimal percent", value: "1050", want: 1050},
		{name: "full rate", value: "10000", want: 10000},
		{name: "negative", value: "-1", wantErr: true},
		{name: "over full rate", value: "10001", wantErr: true},
		{name: "fractional basis points", value: "10.5", wantErr: true},
		{name: "not a number", value: "ten", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateAffiliateCommissionRateBasisPoints(tc.value)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
