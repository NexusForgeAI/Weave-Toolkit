package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// CalculatorTool 计算器工具
type CalculatorTool struct{}

// CalculatorArgs 计算器参数
type CalculatorArgs struct {
	Operation string  `json:"operation"` // add, subtract, multiply, divide
	A         float64 `json:"a"`
	B         float64 `json:"b"`
}

// CalculatorResult 计算结果
type CalculatorResult struct {
	Result float64 `json:"result"`
}

func (ct *CalculatorTool) Name() string {
	return "calculator"
}

func (ct *CalculatorTool) Description() string {
	return "Perform basic arithmetic operations (add, subtract, multiply, divide)"
}

func (ct *CalculatorTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var calcArgs CalculatorArgs
	if err := json.Unmarshal(args, &calcArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	var result float64
	switch calcArgs.Operation {
	case "add":
		result = calcArgs.A + calcArgs.B
	case "subtract":
		result = calcArgs.A - calcArgs.B
	case "multiply":
		result = calcArgs.A * calcArgs.B
	case "divide":
		if calcArgs.B == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = calcArgs.A / calcArgs.B
	default:
		return nil, fmt.Errorf("unsupported operation: %s", calcArgs.Operation)
	}

	return json.Marshal(CalculatorResult{Result: result})
}