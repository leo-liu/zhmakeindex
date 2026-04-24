# 工程化与发布流程

## 文档目的

本文说明 zhmakeindex 在 Go Modules、测试、CI 与发布自动化方面的当前稳定工程化形态，帮助后续任务快速定位“如何验证改动”和“如何产出发行包”。

它不重复程序运行时的处理流水线，而是覆盖仓库维护流程与自动化边界。

## 1. 模块化构建基线

项目已从旧式 GOPATH 形态迁移到 Go Modules：

- `go.mod` 已建立
- module path 为 `github.com/leo-liu/zhmakeindex`
- 依赖版本由模块文件锁定，而不是依赖外部 GOPATH 布局

这带来几个稳定影响：

1. 本地构建、测试与 CI 都以模块模式执行。
2. 跨平台构建不再要求预先布置 GOPATH 风格目录树。
3. 发布自动化可以直接在干净 checkout 上运行 `go build` / `go test` / GoReleaser。

## 2. 测试体系结构

当前仓库的测试分为单元测试、子系统测试和端到端 golden 测试三层。

### 2.1 基础与工具函数测试

- `main_test.go`
  - 覆盖 `stripExt`
  - 覆盖 `checkEncoding`
- `numberedreader_test.go`
  - 覆盖带行号 rune 读取器行为

### 2.2 语义子系统测试

- `pagenumber_test.go`
  - 页码解析、格式化、比较、差值
- `style_test.go`
  - tokenizer、反引号 token、`unquote`、`parseInt`、默认样式
- `sorter_test.go`
  - `IsNumRune`、`DecimalStrcmp`、`Strcmp`、`getStringType`
- `input_test.go`
  - `ScanIndexEntry`、`CompareIndexEntry`、`skipspaces`
- `collator_test.go`
  - 三种中文排序策略的 `RuneCmp`、`IsLetter`、`Group`

### 2.3 端到端 golden 测试

`integration_test.go` 先在测试时构建可执行文件，再以真实 CLI 方式运行样例输入：

- 示例输入：5 组 `examples/*.idx`
- 排序方式：`pinyin` / `stroke` / `radical`
- 样式测试：额外覆盖样式文件场景
- 基线文件：`testdata/*.golden`

当前共有 17 个 golden files，覆盖：

- 5 组输入 × 3 种排序 = 15 组稳定输出
- 2 组样式文件输出基线

这使回归验证不只检查内部函数结果，也检查最终 `.ind` 产物是否保持稳定。

## 3. CI 流程

GitHub Actions CI 定义于 `.github/workflows/ci.yml`。

### 3.1 触发条件

- push 到 `master` / `main`
- pull request 指向 `master` / `main`

### 3.2 矩阵

- Go：1.21 / 1.22 / 1.23
- OS：Linux / macOS / Windows

### 3.3 执行步骤

每个矩阵任务固定执行：

1. `actions/checkout@v4`
2. `actions/setup-go@v5`
3. `go build ./...`
4. `go test ./... -v`
5. `go vet ./...`

### 3.4 稳定意义

CI 当前扮演三类质量门：

- 构建门：确认模块依赖与平台构建链可用
- 回归门：确认测试和 golden 输出未被破坏
- 静态检查门：用 `go vet` 捕捉典型实现错误

因此，如果任务涉及排序、页码、输入解析、样式或平台兼容，CI 结果应被视为仓库级稳定信号。

## 4. Release 自动化

发布流程由 GitHub Actions 与 GoReleaser 共同完成。

### 4.1 触发方式

`.github/workflows/release.yml` 在 push `v*` tag 时触发发布工作流。

### 4.2 工作流职责

发布工作流会：

1. checkout 完整 git 历史
2. 安装 Go 1.23
3. 调用 `goreleaser/goreleaser-action@v6`
4. 执行 `goreleaser release --clean`

工作流使用 GitHub 内置的 `GITHUB_TOKEN` 发布产物。

### 4.3 GoReleaser 产物矩阵

`.goreleaser.yml` 当前配置：

- 二进制名：`zhmakeindex`
- 入口：仓库根目录主包
- 平台：`linux` / `darwin` / `windows`
- 架构：`amd64` / `arm64`
- 归档：默认 `tar.gz`，Windows 改为 `zip`
- 校验文件：`checksums.txt`

### 4.4 版本信息注入

GoReleaser 通过 `ldflags` 注入：

- `main.Version={{.Version}}`
- `main.Revision={{.ShortCommit}}`

这与本地安装脚本的版本注入思路保持一致，使正式发布产物具备可追踪版本元数据。

## 5. Homebrew 分发边界

`packaging/homebrew/zhmakeindex.rb` 提供 Homebrew formula 模板，用于 tap 方式安装。

当前状态应理解为：

- 模板已存在，覆盖 macOS / Linux 与 amd64 / arm64 下载地址模式
- formula 仍需在每次 release 后更新 `version`、`url`、`sha256`
- 仓库中给出了可选的自动化更新思路，但未形成主发布流程的一部分

因此 Homebrew 支持已经有稳定模板约定，但还不是完全自动化分发链的一环。

## 6. 维护时的检索路径

当任务涉及工程化改造时，可按下面路径定位：

- 模块与依赖：`go.mod`
- 单元与集成测试：`*_test.go`、`integration_test.go`
- golden 基线：`testdata/*.golden`
- CI：`.github/workflows/ci.yml`
- 发布：`.github/workflows/release.yml`、`.goreleaser.yml`
- Homebrew 模板：`packaging/homebrew/zhmakeindex.rb`
- 本地历史脚本：`install.sh`、`install.cmd`、`build-dist.cmd`

## 7. 当前工程化边界

- CI 使用的 Go 版本矩阵目前为 1.21 到 1.23，而 `go.mod` 中声明的 Go 版本需与实际支持策略单独核对。
- Release 流程已自动生成跨平台发行包，但未自动回写 Homebrew formula。
- golden 测试覆盖了多个典型输入和排序策略，但其覆盖范围仍以示例集为边界，不代表所有真实索引输入组合。
- 历史本地脚本仍保留在仓库中，因此文档上应将其视为补充入口，而不是唯一发布路径。
