package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatUserLogsHidesUpstreamModelWhenEnabled(t *testing.T) {
	original := common.HideUserLogUpstreamModelEnabled
	t.Cleanup(func() {
		common.HideUserLogUpstreamModelEnabled = original
	})
	common.HideUserLogUpstreamModelEnabled = true

	logs := []*Log{
		{
			Id:          99,
			ChannelName: "secret-channel",
			ModelName:   "public-alias-model",
			Other: common.MapToJsonStr(map[string]interface{}{
				"admin_info":          map[string]interface{}{"use_channel": []int{12}},
				"is_model_mapped":     true,
				"request_path":        "/v1/chat/completions",
				"upstream_model_name": "secret-upstream-model",
			}),
		},
	}

	formatUserLogs(logs, 0)

	require.Len(t, logs, 1)
	assert.Equal(t, "public-alias-model", logs[0].ModelName)
	assert.Empty(t, logs[0].ChannelName)

	other, err := common.StrToMap(logs[0].Other)
	require.NoError(t, err)
	assert.NotContains(t, other, "admin_info")
	assert.NotContains(t, other, "is_model_mapped")
	assert.NotContains(t, other, "upstream_model_name")
	assert.Equal(t, "/v1/chat/completions", other["request_path"])
}

func TestFormatUserLogsKeepsUpstreamModelWhenDisabled(t *testing.T) {
	original := common.HideUserLogUpstreamModelEnabled
	t.Cleanup(func() {
		common.HideUserLogUpstreamModelEnabled = original
	})
	common.HideUserLogUpstreamModelEnabled = false

	logs := []*Log{
		{
			ModelName: "public-alias-model",
			Other: common.MapToJsonStr(map[string]interface{}{
				"is_model_mapped":     true,
				"upstream_model_name": "secret-upstream-model",
			}),
		},
	}

	formatUserLogs(logs, 0)

	other, err := common.StrToMap(logs[0].Other)
	require.NoError(t, err)
	assert.Equal(t, true, other["is_model_mapped"])
	assert.Equal(t, "secret-upstream-model", other["upstream_model_name"])
}

func TestTaskGetAllUserTaskHidesUpstreamModelWhenEnabled(t *testing.T) {
	truncateTables(t)
	original := common.HideUserLogUpstreamModelEnabled
	t.Cleanup(func() {
		common.HideUserLogUpstreamModelEnabled = original
	})
	common.HideUserLogUpstreamModelEnabled = true

	require.NoError(t, DB.Create(&Task{
		TaskID: "task_public",
		UserId: 903,
		Properties: Properties{
			OriginModelName:   "public-alias-model",
			UpstreamModelName: "secret-upstream-model",
		},
	}).Error)

	tasks := TaskGetAllUserTask(903, 0, 10, SyncTaskQueryParams{})

	require.Len(t, tasks, 1)
	assert.Equal(t, "public-alias-model", tasks[0].Properties.OriginModelName)
	assert.Empty(t, tasks[0].Properties.UpstreamModelName)
}

func TestTaskGetAllTasksKeepsUpstreamModelForAdminWhenHideEnabled(t *testing.T) {
	truncateTables(t)
	original := common.HideUserLogUpstreamModelEnabled
	t.Cleanup(func() {
		common.HideUserLogUpstreamModelEnabled = original
	})
	common.HideUserLogUpstreamModelEnabled = true

	require.NoError(t, DB.Create(&Task{
		TaskID: "task_admin",
		UserId: 904,
		Properties: Properties{
			OriginModelName:   "public-alias-model",
			UpstreamModelName: "secret-upstream-model",
		},
	}).Error)

	tasks := TaskGetAllTasks(0, 10, SyncTaskQueryParams{})

	require.Len(t, tasks, 1)
	assert.Equal(t, "secret-upstream-model", tasks[0].Properties.UpstreamModelName)
}
