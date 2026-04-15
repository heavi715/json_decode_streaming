# JSON Decode Streaming（strict_prefix）

一个面向多语言的流式 JSON 修复工具包，用于处理 AI 输出在末尾被截断的场景。

[English README](../README.md)

本项目使用的策略是 `strict_prefix`：

- 仅保留从开头起最长的安全前缀；
- 自动补全缺失的字符串引号（仅值字符串）和容器结束符；
- 绝不猜测缺失的非空值，也不注入新键。

## 支持语言（阶段 1）

- Python：`python/repair_json.py`
- JavaScript：`javascript/repairJson.js`
- Go：`golang/repair_json.go`
- PHP：`php/RepairJson.php`

## 在其他项目中安装

当前发布状态：

- Python（PyPI）：可用
- JavaScript（npm）：可用
- Go module：可用
- PHP（Packagist）：可能有索引延迟；可使用 VCS 兜底安装

远端直接安装：

- Python：`pip install json-decode-streaming`
- JavaScript：`npm install json-decode-streaming`
- Go：`go get github.com/heavi715/json_decode_streaming@v0.1.6`
- PHP（Packagist 已索引时）：`composer require heavi/json-decode-streaming:^0.1`

PHP 兜底安装（VCS，始终可用）：

- `composer config repositories.json_decode_streaming vcs https://github.com/heavi715/json_decode_streaming.git`
- `composer require heavi/json-decode-streaming:v0.1.6`

## 本地构建可分发产物

从源码构建各语言包：

- Python wheel/sdist：
  - `make package-python`
  - 输出：`python/dist/`
- JavaScript npm tarball：
  - `make package-javascript`
  - 输出：`javascript/*.tgz`
- PHP Composer archive：
  - `make package-php`
  - 输出：`php/heavi-json-decode-streaming-*.zip`
- Go module 检查：
  - `make package-go`（检查模块可导入/可构建）

一键构建全部：

- `make package-all`

从本地产物安装：

- Python：`pip install python/dist/*.whl`
- JavaScript：`npm install ./javascript/*.tgz`
- PHP：`composer config repositories.local artifact ./php && composer require heavi/json-decode-streaming:*`

## 示例行为

- `{"a":"b` -> `{"a":"b"}`
- `{"a":[1,2,` -> `{"a":[1,2]}`
- `{"a":"1","b":` -> `{"a":"1"}`
- `{"a":"1","b":"` -> `{"a":"1","b":""}`

## 直接返回解析对象

- Python：`repair_json_strict_prefix(text, return_object=True)` 返回解析对象（`dict`/`list`/primitive），若修复结果为空则返回 `None`。
- Python：`repair_json_strict_prefix_both(text)` 返回 `(repaired_string, parsed_object_or_none)`。
- JavaScript：`repairJsonStrictPrefix(text, true)` 返回解析对象（`object`/`array`/primitive），若修复结果为空则返回 `null`。
- JavaScript：`repairJsonStrictPrefixBoth(text)` 返回 `[repairedString, parsedObjectOrNull]`。
- Go：`RepairJSONStrictPrefixWithOption(text, true)` 返回解析对象（`any`），若修复结果为空则返回 `nil`。
- Go：`RepairJSONStrictPrefixBoth(text)` 返回 `(repairedString, parsedObjectOrNil, err)`。
- PHP：`repair_json_strict_prefix($text, true)` 返回解析后的数组/基础类型，若修复结果为空则返回 `null`。
- PHP：`repair_json_strict_prefix_both($text)` 返回 `[$repaired, $parsedOrNull]`。

## 流式场景下追加内容

当 AI 响应按 chunk 到达时，将已有内容作为 `text`，增量内容作为 append 参数传入：

- Python：`repair_json_strict_prefix(text, return_object=True, append_content=chunk)`
- JavaScript：`repairJsonStrictPrefix(text, true, chunk)`
- Go：`RepairJSONStrictPrefixWithAppendOption(text, chunk, true)`
- PHP：`repair_json_strict_prefix($text, true, $chunk)`

## 运行测试

共享测试向量在 `test/cases.json`。

- 同步检查：
  - `make test-cases-sync`
  - `make test-cases-stats`
- Python：
  - `python3 test/test-python.py`
  - `python3 test/test-fuzz-python.py`
  - `python3 test/test-python-incremental.py`
  - `python3 test/test-ai-stream-python.py`（需要环境变量）
  - `./test/test-ai-stream-curl.sh`（需要环境变量）
- JavaScript：
  - `export PATH="/Users/heavi/.nvm/versions/node/v22.14.0/bin:$PATH"`
  - `node test/test-ai-stream-javascript.js`（需要环境变量）
- Go：
  - `go run -tags aistream test/test-ai-stream-go.go`（需要环境变量）
- PHP：
  - `php test/test-ai-stream-php.php`（需要环境变量）
- JavaScript：
  - `export PATH="/Users/heavi/.nvm/versions/node/v22.14.0/bin:$PATH"`
  - `node test/test-javascript.js`
- Go：
  - `go run test/test-go.go`
- PHP：
  - `php test/test-php.php`

## 运行基准测试

- Python：`python3 test/test-bench-python.py`（输出 `string` 和 `object` 模式）
- Python 增量内存：`python3 test/test-bench-python-incremental-memory.py`
- PHP：`php test/test-bench-php.php`（输出 `string` 和 `object` 模式）
- JavaScript：
  - `export PATH="/Users/heavi/.nvm/versions/node/v22.14.0/bin:$PATH"`
  - `node test/test-bench-javascript.js`（输出 `string` 和 `object` 模式）
- Go：`go run -tags benchgo test/test-bench-go.go`（输出 `string` 和 `object` 模式）

所有基准脚本均支持可选 mode 参数：

- `--mode=string`（Go 对应：`-mode=string`）
- `--mode=object`（Go 对应：`-mode=object`）
- `--mode=both_return`（Go 对应：`-mode=both_return`），用于单次调用 `(repaired, object)` 路径
- 默认值为 `all`

## 覆盖快照

当前共享确定性用例：见 `make test-cases-stats`（读取 `docs/test-cases.md` + `test/cases.json`，并输出各分组数量）。

已覆盖场景分组包括：

- 经典截断示例
- 额外基础闭合与尾部丢弃行为
- 多层嵌套对象/数组闭合
- 边界与前缀策略回滚行为
- 数字 token 与指数/小数边界
- 键和值中的 Unicode/转义处理
- 数组/对象逗号状态修复
- 分隔符不匹配处理
- 根对象完成后与多根截断行为
- 高噪声混合尾部重放基线

## 设计文档

- 规范：`docs/spec.md`
- 测试用例与校验规则：`docs/test-cases.md`
- 增量模式设计（Python）：`docs/incremental-mode.md`
- 标准化发布流程：`docs/publish.md`
- GitHub 自动发布工作流：`.github/workflows/release.yml`

## 复杂度

- 时间复杂度：`O(n)`（单次线性扫描）。
- 空间复杂度：`O(depth)`（容器栈）。

## 说明

- 本项目目标是修复尾部损坏/截断。
- 不尝试修复深度损坏的前缀。
- 流式测试环境变量：`AI_STREAM_API_KEY`，可选 `AI_STREAM_URL`、`AI_STREAM_MODEL`、`AI_STREAM_PROMPT`。
- 流式快照打印：`AI_STREAM_PRINT_SNAPSHOTS=1`（默认），数量上限由 `AI_STREAM_MAX_SNAPSHOTS` 控制（默认 `20`）。