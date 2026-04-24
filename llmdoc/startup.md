# 启动阅读顺序

新会话进入 zhmakeindex 仓库后，按下面顺序读取稳定文档即可建立主要上下文。

1. `llmdoc/overview/project-overview.md`
   - 先了解项目定位、兼容目标、主要特性、技术栈与外部依赖。
2. `llmdoc/architecture/pipeline.md`
   - 这是主执行模型文档，覆盖 `.idx` 输入、解析、合并、排序、页码整理到 `.ind` 输出的完整流水线。
3. `llmdoc/architecture/sorting.md`
   - 当任务涉及中文排序、分组规则、CJK 数据或排序异常时读取。
4. `llmdoc/reference/style-reference.md`
   - 当任务涉及 `.ist/.mst` 样式、输出格式、页码分隔、分组标题或 makeindex 兼容性时读取。

按需补充：

- 若任务涉及构建、安装、版本注入、测试覆盖或数据表生成，再回到代码查看 `install.sh`、`install.cmd`、`build-dist.cmd`、`CJK/maketables.go`。
- 若任务只涉及局部行为，仍应至少读完前 2 项，再进入代码。