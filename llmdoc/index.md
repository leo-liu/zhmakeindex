# zhmakeindex llmdoc 索引

本目录保存 zhmakeindex 的稳定文档。项目是一个用 Go 编写、面向中文用户的 makeindex 替代品，重点在 Unicode 中文排序、多编码 I/O、makeindex 样式兼容，以及近期开启的模块化工程与自动化发布流程。

## overview/

- `llmdoc/overview/project-overview.md`：项目定位、核心能力、技术栈、依赖、测试与 CI/CD 形态、许可证与当前工程特征。

## architecture/

- `llmdoc/architecture/pipeline.md`：主处理流水线，从命令行与样式初始化，到 `.idx` 解析、条目归并、排序、页码区间整理与 `.ind` 渲染。
- `llmdoc/architecture/sorting.md`：中文排序子系统，说明 `IndexCollator` 策略接口、三种排序方式、`Strcmp` 比较层次，以及 CJK 静态数据表的来源与约束。

## guides/

- `llmdoc/guides/engineering-and-release.md`：Go Modules、测试体系、CI 矩阵、GoReleaser 发布流程与 Homebrew 模板边界。

## reference/

- `llmdoc/reference/style-reference.md`：样式文件关键字总表，包含输入样式、输出样式、默认值、兼容别名、已知未生效项与兼容性说明。

## 代码入口提示

- CLI 与编码初始化：`main.go`
- 输入解析：`input.go`、`numberedreader.go`
- 排序与页码处理：`sorter.go`、`reading_collator.go`、`stroke_collator.go`、`radical_collator.go`、`pagenumber.go`
- 样式系统：`style.go`
- 输出渲染：`output.go`
- TeX 样式查找：`kpathsea/kpathsea.go`
- CJK 数据生成：`CJK/maketables.go`
- 测试总览：`*_test.go`、`integration_test.go`、`testdata/*.golden`
- 工程自动化：`.github/workflows/ci.yml`、`.github/workflows/release.yml`、`.goreleaser.yml`

## 当前文档边界

现有文档优先覆盖项目定位、执行流水线、排序系统、样式契约，以及近期工程化改造后的验证与发布流程。若未来任务频繁涉及 Homebrew 发布细节、测试数据维护策略或模块兼容策略，再考虑继续细分文档。
