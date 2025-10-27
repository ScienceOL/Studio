package sandbox

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TemplateTransformer 定义模板转换器接口
type TemplateTransformer interface {
	GetRunnerScript() string
	GetPreloadScript() string
}

// BaseTemplateTransformer 基础模板转换器
type BaseTemplateTransformer struct {
	CodePlaceholder   string
	InputsPlaceholder string
	ResultTag         string
}

// NewBaseTemplateTransformer 创建新的基础模板转换器
func NewBaseTemplateTransformer() *BaseTemplateTransformer {
	return &BaseTemplateTransformer{
		CodePlaceholder:   "{{code}}",
		InputsPlaceholder: "{{inputs}}",
		ResultTag:         "<<RESULT>>",
	}
}

// TransformCaller 转换代码为 runner
func (t *BaseTemplateTransformer) TransformCaller(code string, inputs map[string]any, transformer TemplateTransformer) (string, string, error) {
	runnerScript, err := t.AssembleRunnerScript(code, inputs, transformer)
	if err != nil {
		return "", "", err
	}
	preloadScript := transformer.GetPreloadScript()
	return runnerScript, preloadScript, nil
}

// ExtractResultStrFromResponse 从响应中提取结果字符串
func (t *BaseTemplateTransformer) ExtractResultStrFromResponse(response string) (string, error) {
	pattern := fmt.Sprintf(`%s(.*)%s`, regexp.QuoteMeta(t.ResultTag), regexp.QuoteMeta(t.ResultTag))
	re := regexp.MustCompile(`(?s)` + pattern) // (?s) 等价于 Python 的 re.DOTALL

	matches := re.FindStringSubmatch(response)
	if len(matches) < 2 {
		responsePreview := response
		if len(response) > 200 {
			responsePreview = response[:200] + "..."
		}
		return "", fmt.Errorf("failed to parse result: no result tag found in response. Response: %s", responsePreview)
	}

	return matches[1], nil
}

// TransformResponse 转换响应为字典
func (t *BaseTemplateTransformer) TransformResponse(response string) (map[string]any, error) {
	resultStr, err := t.ExtractResultStrFromResponse(response)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %s", err.Error())
	}

	// 后处理结果，转换科学计数法字符串为数字
	processedResult := t.postProcessResult(result)
	return processedResult, nil
}

// postProcessResult 后处理结果，转换科学计数法字符串为数字
func (t *BaseTemplateTransformer) postProcessResult(result map[string]any) map[string]any {
	return t.convertScientificNotation(result).(map[string]any)
}

// convertScientificNotation 转换科学计数法
func (t *BaseTemplateTransformer) convertScientificNotation(value any) any {
	switch v := value.(type) {
	case string:
		// 检查字符串是否像科学计数法
		matched, _ := regexp.MatchString(`^-?\d+\.?\d*e[+-]\d+$`, strings.ToLower(v))
		if matched {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
		return v
	case map[string]any:
		result := make(map[string]any)
		for k, val := range v {
			result[k] = t.convertScientificNotation(val)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = t.convertScientificNotation(val)
		}
		return result
	default:
		return v
	}
}

// SerializeInputs 序列化输入
func (t *BaseTemplateTransformer) SerializeInputs(inputs map[string]any) (string, error) {
	inputsJSON, err := json.Marshal(inputs)
	if err != nil {
		return "", err
	}

	inputBase64Encoded := base64.StdEncoding.EncodeToString(inputsJSON)
	return inputBase64Encoded, nil
}

// AssembleRunnerScript 组装运行器脚本
func (t *BaseTemplateTransformer) AssembleRunnerScript(code string, inputs map[string]any, transformer TemplateTransformer) (string, error) {
	script := transformer.GetRunnerScript()
	script = strings.ReplaceAll(script, t.CodePlaceholder, code)

	inputsStr, err := t.SerializeInputs(inputs)
	if err != nil {
		return "", err
	}

	script = strings.ReplaceAll(script, t.InputsPlaceholder, inputsStr)
	return script, nil
}

// Python3TemplateTransformer Python3 模板转换器（严格函数调用方式）
type Python3TemplateTransformer struct {
	*BaseTemplateTransformer
}

// NewPython3TemplateTransformer 创建 Python3 模板转换器
func NewPython3TemplateTransformer() TemplateTransformer {
	return &Python3TemplateTransformer{
		BaseTemplateTransformer: NewBaseTemplateTransformer(),
	}
}

// GetRunnerScript 获取运行器脚本（严格函数调用方式）
func (p *Python3TemplateTransformer) GetRunnerScript() string {
	return `
# declare main function
{{code}}

import json
from base64 import b64decode

# decode and prepare input dict
inputs_obj = json.loads(b64decode('{{inputs}}').decode('utf-8'))

# execute main function
output_obj = main(**inputs_obj)

# convert output to json and print
output_json = json.dumps(output_obj, indent=4)
result = f'''<<RESULT>>{output_json}<<RESULT>>'''
print(result)
`
}

// GetPreloadScript 获取预加载脚本
func (p *Python3TemplateTransformer) GetPreloadScript() string {
	return ""
}
