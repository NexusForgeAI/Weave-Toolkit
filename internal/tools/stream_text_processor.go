package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// StreamTextProcessor 流式文本处理工具
type StreamTextProcessor struct{}

// StreamTextArgs 流式文本处理参数
type StreamTextArgs struct {
	Text      string `json:"text"`
	Operation string `json:"operation"` // split, reverse, count, analyze
}

// StreamTextResult 流式文本处理结果
type StreamTextResult struct {
	OriginalText string      `json:"original_text"`
	Result       interface{} `json:"result"`
	Operation    string      `json:"operation"`
}

// parseArguments 统一参数解析函数
func parseArguments(args json.RawMessage) (StreamTextArgs, error) {
	var textArgs StreamTextArgs

	// 首先尝试解析为 map
	var rawArgs map[string]interface{}
	if err := json.Unmarshal(args, &rawArgs); err != nil {
		return textArgs, fmt.Errorf("invalid JSON format: %v", err)
	}

	// 提取参数
	if textVal, ok := rawArgs["text"].(string); ok {
		textArgs.Text = textVal
	}

	if opVal, ok := rawArgs["operation"].(string); ok {
		textArgs.Operation = opVal
	} else {
		// 如果没有提供 operation 参数，使用默认值
		textArgs.Operation = "analyze"
	}

	// 验证参数
	if textArgs.Text == "" {
		return textArgs, fmt.Errorf("text parameter is required")
	}

	return textArgs, nil
}

func (stp *StreamTextProcessor) Name() string {
	return "stream_text_processor"
}

func (stp *StreamTextProcessor) Description() string {
	return "Process text with streaming output (split, reverse, count, analyze)"
}

func (stp *StreamTextProcessor) Category() ToolCategory {
	return CategoryUtility
}

func (stp *StreamTextProcessor) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	// 使用统一的参数解析函数
	textArgs, err := parseArguments(args)
	if err != nil {
		return nil, err
	}

	var result interface{}
	switch textArgs.Operation {
	case "split":
		result = strings.Fields(textArgs.Text)
	case "reverse":
		runes := []rune(textArgs.Text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		result = string(runes)
	case "count":
		result = map[string]int{
			"characters": len(textArgs.Text),
			"words":      len(strings.Fields(textArgs.Text)),
			"lines":      len(strings.Split(textArgs.Text, "\n")),
		}
	case "analyze":
		result = map[string]interface{}{
			"length":        len(textArgs.Text),
			"word_count":    len(strings.Fields(textArgs.Text)),
			"line_count":    len(strings.Split(textArgs.Text, "\n")),
			"has_uppercase": strings.ToLower(textArgs.Text) != textArgs.Text,
			"has_lowercase": strings.ToUpper(textArgs.Text) != textArgs.Text,
		}
	default:
		return nil, fmt.Errorf("unsupported operation: %s", textArgs.Operation)
	}

	return json.Marshal(StreamTextResult{
		OriginalText: textArgs.Text,
		Result:       result,
		Operation:    textArgs.Operation,
	})
}

// ExecuteStream 流式执行文本处理
func (stp *StreamTextProcessor) ExecuteStream(ctx context.Context, args json.RawMessage, callback func(content string, index int)) (json.RawMessage, error) {
	// 使用统一的参数解析函数
	textArgs, err := parseArguments(args)
	if err != nil {
		return nil, err
	}

	// 流式处理流程
	callback("开始文本处理...", 0)
	time.Sleep(100 * time.Millisecond)

	callback(fmt.Sprintf("输入文本长度: %d 字符", len(textArgs.Text)), 1)
	time.Sleep(100 * time.Millisecond)

	callback(fmt.Sprintf("处理操作: %s", textArgs.Operation), 2)
	time.Sleep(100 * time.Millisecond)

	// 使用普通调用获取结果，然后流式展示处理过程
	result, err := stp.Execute(ctx, args)
	if err != nil {
		callback(fmt.Sprintf("处理失败: %s", err.Error()), 3)
		return nil, err
	}

	// 解析结果用于流式展示
	var streamResult StreamTextResult
	if err := json.Unmarshal(result, &streamResult); err != nil {
		callback("结果解析失败", 3)
		return nil, err
	}

	// 根据操作类型展示不同的处理过程
	switch textArgs.Operation {
	case "split":
		callback("正在分割文本...", 3)
		if words, ok := streamResult.Result.([]interface{}); ok {
			for i, word := range words {
				callback(fmt.Sprintf("单词 %d: %s", i+1, word), 4+i)
				time.Sleep(30 * time.Millisecond)
			}
		}
		callback("文本分割完成", 4+len(streamResult.Result.([]interface{})))

	case "reverse":
		callback("正在反转文本...", 3)
		time.Sleep(200 * time.Millisecond)
		callback("文本反转完成", 4)
		callback(fmt.Sprintf("结果: %s", streamResult.Result), 5)

	case "count":
		callback("正在统计文本信息...", 3)
		time.Sleep(200 * time.Millisecond)
		if counts, ok := streamResult.Result.(map[string]interface{}); ok {
			callback(fmt.Sprintf("统计完成: %v 字符, %v 单词, %v 行",
				counts["characters"], counts["words"], counts["lines"]), 4)
		}

	case "analyze":
		callback("正在分析文本特征...", 3)
		time.Sleep(200 * time.Millisecond)
		callback("文本分析完成", 4)

	default:
		callback(fmt.Sprintf("错误：不支持的操作类型 %s", textArgs.Operation), 3)
		return nil, fmt.Errorf("unsupported operation: %s", textArgs.Operation)
	}

	callback("处理完成！", 100)

	return result, nil
}
