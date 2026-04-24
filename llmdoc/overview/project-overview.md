# 项目总览

## 项目定位

zhmakeindex 是一个用 Go 实现的 `makeindex` 替代品，面向中文 LaTeX 索引生成场景。它保留了 makeindex 的输入语法、样式文件模型和常见命令行工作方式，同时补上标准 makeindex 对中文与 Unicode 支持不足的问题。

项目的稳定定位可以概括为两点：

1. 兼容 makeindex 的索引处理模型，而不是重新发明一套索引格式。
2. 针对中文文献环境增强排序与编码能力，支持 Unicode 汉字排序和传统中文工作流中的多种历史编码。

## 核心功能

### 中文排序

项目内建三种中文排序/分组方式，由 `-z` 选择：

- `pinyin` / `reading`：按拼音排序，汉字并入对应拼音首字母组。
- `stroke` / `bihua`：按总笔画数、笔顺串排序，并按笔画数分组。
- `radical` / `bushou`：按部首、余画排序，并按 214 部首分组。

### makeindex 兼容输入与输出

- 默认读取 makeindex 风格的 `.idx` 记录，关键字默认为 `\indexentry`。
- 支持层级条目、`key@text`、`|encap`、`|(` / `|)` 区间语法。
- 输出 `.ind` 时复用 makeindex 风格的样式参数体系。
- 样式文件缺省扩展名为 `.ist`；当输入只有一个文件且同名 `.mst` 存在时，会自动采用该样式文件。

### 多编码 I/O

输入输出编码由 `-enc` 控制，样式文件编码由 `-senc` 控制，当前支持：

- `utf-8`
- `utf-16`
- `gb18030`
- `gbk`
- `big5`

内部处理统一先解码到 Unicode rune，再做解析、排序和输出拼装，因此编码兼容主要集中在 I/O 边界。

### 页码排序与区间处理

- 支持阿拉伯数字、大小写罗马数字、大小写单字母页码。
- 支持复合页码，默认分隔符为 `-`，也可由样式关键字 `page_compositor` 改写。
- 支持显式区间与自动区间合并。
- `-strict` 用于严格区分不同 `encap` 命令下的页码；`-r` 用于关闭自动连续页合并。

### 与 TeX 发行版集成

样式文件查找先检查当前路径，再通过 `kpsewhich` 搜索 TEXMF 树。当前集成方式只依赖命令行工具 `kpsewhich`，不依赖 kpathsea 动态库。

## 技术栈

### 语言与工程形态

- 语言：Go
- 结构：单二进制 CLI 程序
- 模块管理：Go Modules
- module path：`github.com/leo-liu/zhmakeindex`

### 主要源码分层

- `main.go`：命令行、编码初始化、日志、主流程编排
- `style.go`：样式文件解析，分别构造 `InputStyle` 与 `OutputStyle`
- `input.go`：`.idx` 解析、语法错误恢复、索引条目归并
- `numberedreader.go`：带行号的 rune 级读取器
- `sorter.go`：整体排序、分组与页码整理入口
- `reading_collator.go` / `stroke_collator.go` / `radical_collator.go`：三种中文排序策略
- `pagenumber.go`：页码解析、比较、格式化
- `output.go`：按样式渲染 `.ind`
- `kpathsea/kpathsea.go`：通过 `kpsewhich` 查找样式文件
- `CJK/*.go`：运行期使用的静态汉字排序数据表
- `CJK/maketables.go`：构建期数据生成器

## 主要依赖

### Go 依赖

- `golang.org/x/text`
  - 用于编码器/解码器与 `transform.Reader/Writer`
- `github.com/yasushi-saito/rbtree`
  - 用于输入阶段的红黑树集合，实现条目去重、排序遍历与父级补全

### 外部命令与数据来源

- `kpsewhich`
  - 用于样式文件搜索，连接 TeX 发行版的 TEXMF 文件树
- Unicode Unihan 数据
  - `CJK/maketables.go` 会下载 `Unihan.zip` 生成拼音、笔画与部首数据表
- `CJK/sunwb_strokeorder.txt`
  - 笔顺表生成时使用的本地数据源

## 构建、测试与发布形态

### 本地构建与安装

- `go.mod` 已建立，项目现在以 Go Modules 方式构建与测试。
- `install.sh` / `install.cmd`：本机安装，并通过 `-ldflags` 注入 `main.Version` 与 `main.Revision`
- `build-dist.cmd`：保留的历史发行脚本，体现模块化改造前的本地分发方式

### 测试体系

仓库已建立覆盖核心子系统的 Go 测试体系：

- `main_test.go`：`stripExt`、`checkEncoding`
- `pagenumber_test.go`：页码解析、格式化、比较、差值
- `numberedreader_test.go`：行号读取器
- `style_test.go`：样式 tokenizer、反引号 token、`unquote`、`parseInt`、默认样式
- `sorter_test.go`：`IsNumRune`、`DecimalStrcmp`、`Strcmp`、`getStringType`
- `input_test.go`：`ScanIndexEntry`、`CompareIndexEntry`、`skipspaces`
- `collator_test.go`：三种中文排序策略的 `RuneCmp`、`IsLetter`、`Group`
- `integration_test.go`：端到端 golden file 测试，覆盖 5 组示例输入 × 3 种排序方式，以及样式文件测试
- `testdata/*.golden`：17 个 golden files，作为稳定输出基线

这意味着项目已经从“局部样式解析测试”升级为“单元测试 + 排序策略测试 + 输入解析测试 + 集成 golden 测试”的完整回归体系。

### CI

仓库已配置 GitHub Actions 持续集成：

- 工作流：`.github/workflows/ci.yml`
- 触发条件：推送或 Pull Request 到 `master` / `main`
- 矩阵：Go 1.21 / 1.22 / 1.23 × Linux / macOS / Windows
- 执行内容：`go build ./...`、`go test ./... -v`、`go vet ./...`

CI 的引入把跨平台构建、测试与静态检查从本地手工执行提升为仓库级自动验证。

### Release 自动化

仓库已建立 tag 驱动的发布流程：

- GoReleaser 配置：`.goreleaser.yml`
  - 交叉编译 `linux` / `darwin` / `windows` × `amd64` / `arm64`
  - 注入 `main.Version` 与 `main.Revision`
  - 生成归档包与 `checksums.txt`
- 发布工作流：`.github/workflows/release.yml`
  - 由 `v*` tag 触发
  - 使用 GoReleaser 执行正式发布
- Homebrew 模板：`packaging/homebrew/zhmakeindex.rb`
  - 提供 tap 方式安装所需的 formula 模板
  - 当前仍需在 release 后填充版本、下载 URL 与 SHA256

因此，项目当前已具备从 tag 到跨平台二进制产物的自动化发布骨架，但 Homebrew formula 仍属于半自动维护环节。

## 质量与已知工程特征

- 测试覆盖已从单一 `style_test.go` 扩展到命令行辅助函数、输入解析、页码系统、排序系统、三种 collator 和端到端 golden 测试。
- CI 会在多 Go 版本、多操作系统上执行构建、测试和 `go vet`，工程质量门槛显著高于早期本地脚本驱动阶段。
- 发布链路已经标准化为 GitHub Actions + GoReleaser，跨平台产物生成不再依赖手工执行历史脚本。
- CJK 排序依赖静态生成数据表，运行期查表性能好，但数据更新依赖生成脚本与上游 Unicode 数据。
- 输出阶段当前显式支持到三级索引；更深层级只记录日志并忽略。
- `packaging/homebrew/zhmakeindex.rb` 目前是模板而不是自动回写产物，说明包管理器分发还未完全闭环。

## 许可证

项目采用 LaTeX Project Public License (LPPL) 1.3 或更新版本。版权所有者为刘海洋 (Leo Liu) <leoliu.pku@gmail.com>，维护状态为 `maintained`。
