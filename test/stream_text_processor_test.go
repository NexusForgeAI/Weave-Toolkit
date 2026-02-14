package test

import (
	"context"
	"encoding/json"
	"testing"

	"Weave-Toolkit/internal/tools"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamTextProcessor(t *testing.T) {
	processor := &tools.StreamTextProcessor{}

	tests := []struct {
		name      string
		text      string
		operation string
		expected  interface{}
		hasError  bool
	}{
		{
			name:      "文本分割",
			text:      "hello world test",
			operation: "split",
			expected:  []interface{}{"hello", "world", "test"},
		},
		{
			name:      "文本反转",
			text:      "hello",
			operation: "reverse",
			expected:  "olleh",
		},
		{
			name:      "字符计数",
			text:      "hello world",
			operation: "count",
			expected: map[string]interface{}{
				"characters": float64(11),
				"words":      float64(2),
				"lines":      float64(1),
			},
		},
		{
			name:      "文本分析",
			text:      "Hello World!",
			operation: "analyze",
			expected: map[string]interface{}{
				"length":        float64(12),
				"word_count":    float64(2),
				"line_count":    float64(1),
				"has_uppercase": true,
				"has_lowercase": true,
			},
		},
		{
			name:      "无效操作",
			text:      "test",
			operation: "invalid",
			hasError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]interface{}{
				"text":      tt.text,
				"operation": tt.operation,
			}

			argsJSON, err := json.Marshal(args)
			require.NoError(t, err)

			result, err := processor.Execute(context.Background(), argsJSON)

			if tt.hasError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			var textResult tools.StreamTextResult
			err = json.Unmarshal(result, &textResult)
			require.NoError(t, err)

			assert.Equal(t, tt.text, textResult.OriginalText)
			assert.Equal(t, tt.operation, textResult.Operation)

			// 使用JSON序列化比较结果，避免类型不匹配问题
			expectedJSON, _ := json.Marshal(tt.expected)
			actualJSON, _ := json.Marshal(textResult.Result)
			assert.JSONEq(t, string(expectedJSON), string(actualJSON))
		})
	}
}

func TestStreamTextProcessorDefaultOperation(t *testing.T) {
	processor := &tools.StreamTextProcessor{}

	// 测试默认操作（analyze）
	args := map[string]interface{}{
		"text": "Test Text",
	}

	argsJSON, err := json.Marshal(args)
	require.NoError(t, err)

	result, err := processor.Execute(context.Background(), argsJSON)
	require.NoError(t, err)

	var textResult tools.StreamTextResult
	err = json.Unmarshal(result, &textResult)
	require.NoError(t, err)

	assert.Equal(t, "Test Text", textResult.OriginalText)
	assert.Equal(t, "analyze", textResult.Operation)

	// 验证分析结果
	analysis, ok := textResult.Result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(9), analysis["length"])
	assert.Equal(t, float64(2), analysis["word_count"])
}
