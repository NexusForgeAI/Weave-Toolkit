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
	Operation string    `json:"operation"` // add, subtract, multiply, divide
	A         float64   `json:"a"`
	B         float64   `json:"b"`
	Operands  []float64 `json:"operands"`
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

func (ct *CalculatorTool) Category() ToolCategory {
	return CategoryMath
}

func (ct *CalculatorTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var calcArgs CalculatorArgs
	if err := json.Unmarshal(args, &calcArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	// 支持两种参数格式：{"a":10,"b":20} 或 {"operands":[10,20]}
	var a, b float64
	if len(calcArgs.Operands) >= 2 {
		a = calcArgs.Operands[0]
		b = calcArgs.Operands[1]
	} else {
		a = calcArgs.A
		b = calcArgs.B
	}

	var result float64
	switch calcArgs.Operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return nil, fmt.Errorf("unsupported operation: %s", calcArgs.Operation)
	}

	return json.Marshal(CalculatorResult{Result: result})
}
