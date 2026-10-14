package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/waffo-com/waffo-go/core"
)

func TestWaffoTerminalFailureStatus(t *testing.T) {
	testCases := []struct {
		status string
		want   bool
	}{
		{status: core.OrderStatusPayInProgress, want: false},
		{status: core.OrderStatusAuthorizationRequired, want: false},
		{status: core.OrderStatusAuthedWaitingCapture, want: false},
		{status: core.OrderStatusPaySuccess, want: false},
		{status: core.OrderStatusOrderClose, want: true},
		{status: "UNKNOWN", want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.status, func(t *testing.T) {
			assert.Equal(t, tc.want, isWaffoTerminalFailureStatus(tc.status))
		})
	}
}
