package test

import (
	"context"
	"encoding/json"
	"testing"

	"Weave-Toolkit/internal/tools"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculatorTool(t *testing.T) {
	calculator := &tools.CalculatorTool{}

	tests := []struct {
		name     string
		args     tools.CalculatorArgs
		expected float64
		hasError bool
	}{
		{
			name: "加法运算",
			args: tools.CalculatorArgs{
				Operation: "add",
				A:         10,
				B:         20,
			},
			expected: 30,
		},
		{
			name: "减法运算",
			args: tools.CalculatorArgs{
				Operation: "subtract",
				A:         50,
				B:         30,
			},
			expected: 20,
		},
		{
			name: "乘法运算",
			args: tools.CalculatorArgs{
				Operation: "multiply",
				A:         5,
				B:         6,
			},
			expected: 30,
		},
		{
			name: "除法运算",
			args: tools.CalculatorArgs{
				Operation: "divide",
				A:         100,
				B:         4,
			},
			expected: 25,
		},
		{
			name: "除零错误",
			args: tools.CalculatorArgs{
				Operation: "divide",
				A:         10,
				B:         0,
			},
			hasError: true,
		},
		{
			name: "无效操作",
			args: tools.CalculatorArgs{
				Operation: "invalid",
				A:         10,
				B:         20,
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsJSON, err := json.Marshal(tt.args)
			require.NoError(t, err)

			result, err := calculator.Execute(context.Background(), argsJSON)

			if tt.hasError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			var calcResult tools.CalculatorResult
			err = json.Unmarshal(result, &calcResult)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, calcResult.Result)
		})
	}
}

func TestCalculatorToolWithOperands(t *testing.T) {
	calculator := &tools.CalculatorTool{}

	tests := []struct {
		name      string
		operands  []float64
		operation string
		expected  float64
	}{
		{
			name:      "使用操作数数组进行加法",
			operands:  []float64{15, 25},
			operation: "add",
			expected:  40,
		},
		{
			name:      "使用操作数数组进行乘法",
			operands:  []float64{3, 7},
			operation: "multiply",
			expected:  21,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tools.CalculatorArgs{
				Operation: tt.operation,
				Operands:  tt.operands,
			}

			argsJSON, err := json.Marshal(args)
			require.NoError(t, err)

			result, err := calculator.Execute(context.Background(), argsJSON)
			require.NoError(t, err)

			var calcResult tools.CalculatorResult
			err = json.Unmarshal(result, &calcResult)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, calcResult.Result)
		})
	}
}
