package console_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateActivityBarItems(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty string is valid (noop)",
			input:   "",
			wantErr: false,
		},
		{
			name:    "valid activity items",
			input:   `[{"id":1,"title":"充值抽奖活动进行中！","tag":"抽奖特惠","tagColor":"orange","link":"https://gift.inkwan.app/lottery","enabled":true},{"id":2,"title":"查看模型价格表","tag":"模型价格","tagColor":"green","link":"/pricing","enabled":true}]`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `[{"id":1, "title":]`,
			wantErr: true,
			errMsg:  "Invalid activity bar JSON",
		},
		{
			name:    "missing title",
			input:   `[{"id":1,"tag":"抽奖特惠","link":"https://example.com"}]`,
			wantErr: true,
			errMsg:  "requires a title",
		},
		{
			name:    "empty title",
			input:   `[{"id":1,"title":"   ","tag":"抽奖特惠"}]`,
			wantErr: true,
			errMsg:  "requires a title",
		},
		{
			name:    "dangerous content in title",
			input:   `[{"id":1,"title":"<script>alert(1)</script>"}]`,
			wantErr: true,
			errMsg:  "contains disallowed content",
		},
		{
			name:    "dangerous content in link",
			input:   `[{"id":1,"title":"活动","link":"javascript:alert(1)"}]`,
			wantErr: true,
			errMsg:  "contains disallowed content",
		},
		{
			name:    "invalid external url",
			input:   `[{"id":1,"title":"活动","link":"ftp://bad-url"}]`,
			wantErr: true,
			errMsg:  "invalid URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConsoleSettings(tt.input, "ActivityBarItems")
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
