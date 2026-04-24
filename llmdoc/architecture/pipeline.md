# 处理流水线架构

## 文档目的

本文描述 zhmakeindex 从命令行启动到输出 `.ind` 的完整处理流水线，重点解释数据在各阶段的形态变化、核心结构之间的因果关系，以及几个需要长期保持的系统不变量。

主入口链路对应 `main.go` 中的固定顺序：

1. `option.parse()`
2. `NewStyles(&option.StyleOptions)`
3. `NewInputIndex(&option.InputOptions, instyle)`
4. `NewOutputIndex(in, &option.OutputOptions, outstyle)`
5. `out.Output(&option.OutputOptions)`

这是一条典型的“读取配置 → 解析输入 → 组织语义结构 → 渲染输出”的编译式流水线。

## 1. 入口与配置初始化

### 1.1 选项对象的职责拆分

`Options` 聚合三组子配置：

- `InputOptions`
  - 输入源、输入解码器、是否从 stdin 读取、是否裁剪条目首尾空白
- `OutputOptions`
  - 输出编码器、输出文件、中文排序方式、页码严格模式、自动区间合并开关
- `StyleOptions`
  - 样式文件路径、样式文件解码器

此外还有日志、静默模式和 CPU profile 等外围运行控制项。

### 1.2 命令行解析阶段完成的推导

`Options.parse()` 不只是 `flag.Parse()` 的薄封装，它还负责把 makeindex 风格 CLI 的若干隐式约定提前转成显式配置：

- 清理所有输入路径。
- 禁止“同时给文件参数和 `-i` 标准输入”的冲突组合。
- 若未指定 `-o`，则用第一个输入文件主名推导 `.ind`。
- 若未指定 `-t`，则用第一个输入文件主名推导 `.ilg`。
- 若未指定 `-s` 且只有一个输入文件，则尝试同名 `.mst`。
- 初始化索引文件编码器/解码器。
- 初始化样式文件解码器。

这一步的结果是：后续各阶段不必再次推导默认路径或编码，只消费已经归一化过的配置。

### 1.3 编码层的边界

编码只在 I/O 边界生效：

- 读 `.idx` 时，字节流先通过 `transform.NewReader(..., decoder)` 解码。
- 读样式时，样式文件单独通过 `style_decoder` 解码。
- 写 `.ind` 时，字符流通过 `transform.NewWriter(..., encoder)` 编码。

系统内部的解析、排序、分组和模板拼装统一工作在 Unicode rune 上。这是多编码支持成立的关键不变量：

- 内核不关心外部字节编码。
- 上层语义处理只见到统一的 Unicode 文本。

## 2. 样式层：把 makeindex 契约拆为输入语法和输出模板

`NewStyles()` 返回两个对象：

- `InputStyle`
- `OutputStyle`

这样拆分的意义是：

- 输入阶段只依赖 `.idx` 语法符号约定，如 `keyword`、`level`、`actual`、`encap`。
- 输出阶段只依赖 `.ind` 渲染模板和页码展示规则，如 `item_0`、`delim_r`、`suffix_2p`。

样式加载流程是：

1. 先创建输入/输出样式默认值。
2. 若用户未提供样式文件，直接返回默认值。
3. 若文件名无扩展名，补 `.ist`。
4. 先直接查当前路径，找不到时交给 `kpathsea.FindFile()`，后者通过 `kpsewhich` 在 TEXMF 树中搜索。
5. 读取样式 token，对应关键字覆盖默认值。
6. 未知关键字只记日志，不中断执行。

因此，样式系统的核心语义是“默认值 + 兼容式覆盖”，而不是“严格声明全部参数”。

## 3. 输入层：`.idx` 文本到 `IndexEntry`

### 3.1 输入源

`NewInputIndex()` 支持两种来源：

- `-i` 时从 `os.Stdin` 读取。
- 否则逐个打开输入文件读取；若传入文件名无扩展且不存在，会自动补 `.idx` 再试一次。

多个输入文件会被读入同一个红黑树集合中，因此“多文件合并”是输入层的原生能力。

### 3.2 带行号的 rune 读取

每个输入文件在真正解析前都会包成：

1. `transform.NewReader(idxfile, option.decoder)`
2. `NewNumberdReader(...)`

`NumberdReader` 提供四个关键能力：

- `ReadRune()`：逐 rune 读取并维护行号
- `UnreadRune()`：支持回退一个 rune
- `SkipLine()`：语法错误后直接跳到下一行
- `Line()`：暴露当前行号用于日志

这使得 `input.go` 能实现两个关键特性：

- 以 rune 为单位处理转义、引号、嵌套花括号和层级分隔符
- 在遇到单条索引语法错误时按行恢复，而不是整个输入终止

### 3.3 单条记录扫描

`ScanIndexEntry()` 负责把一条 makeindex 风格记录解析成结构化索引项。逻辑分两段：

#### 第一段：解析索引主体

从第一个 `{...}` 参数中解析：

- 多级层次分隔 `!`
- 排序键 / 显示文本分离 `@`
- `|encap`
- `|(` 与 `|)` 区间标记
- 转义与引号
- 内嵌花括号深度

产出结果写入：

- `entry.level []IndexEntryLevel`
- `page.encap`
- `page.rangetype`
- `entry.input`（原始条目文本快照，用于日志）

#### 第二段：解析页码参数

从第二个 `{...}` 参数中读取页码文本，再调用 `scanPage()` 生成 `[]PageNumber`，最后挂回 `Page.numbers`。

### 3.4 输入阶段核心数据结构

#### `IndexEntry`

输入层的基础语义单元：

- `input string`：原始条目文本，用于问题定位
- `level []IndexEntryLevel`：多级条目路径
- `pagelist []*Page`：尚未排序/合并的页码记录

#### `IndexEntryLevel`

每一级索引包含两部分：

- `key`：排序比较使用
- `text`：输出显示使用

这保证了 makeindex 的“隐藏排序键、单独显示文本”模型在内部是稳定保存的。

#### `Page`

页码不是字符串，而是结构化对象：

- `numbers []PageNumber`
- `compositor string`
- `encap string`
- `rangetype RangeType`

这为后续页码排序、区间构造和输出格式化提供了足够语义。

## 4. 输入归并层：红黑树去重与父级补全

`NewInputIndex()` 不直接把解析结果 append 到切片，而是先进入 `rbtree.Tree`。

### 4.1 去重规则

树键比较使用 `CompareIndexEntry()`，比较顺序为：

1. 逐级比较 `level[i].key`
2. 若相同，再比较 `level[i].text`
3. 前缀完全相同则短层级在前

因此“逻辑上同一条索引项”会被合并到同一个树节点中，其 `pagelist` 被追加合并。

### 4.2 父级补全

当插入一个新条目时，系统还会继续向上构造其父级条目：

- 父级保留层级路径
- 父级 `pagelist` 为空
- 若某级父条目已存在则停止

例如读到 `A!B!C`，系统会确保 `A!B` 与 `A` 也出现在集合中。

这个不变量很重要：

- 输出时子条目不要求父级在原始 `.idx` 中显式出现。
- 输入层会自动补齐层级骨架，保证后续分组/渲染结构完整。

### 4.3 规范化输入结果

红黑树最终按排序遍历转成：

- `type InputIndex []IndexEntry`

此时得到的并不是最终输出顺序，而是已经“去重、合并、补父级”的规范化输入集合。

## 5. 排序与组织层：`InputIndex` 到 `OutputIndex`

`NewOutputIndex()` 只做三件事：

1. 根据 `-z` 创建 `IndexSorter`
2. 调用 `SortIndex()`
3. 将 `style` 与 `option` 挂回结果对象

真正的业务转换都在 `SortIndex()` 内完成。

### 5.1 总体顺序

`SortIndex()` 的固定顺序是：

1. `InitGroups()` 初始化所有可能的输出组
2. 对整个 `InputIndex` 做条目排序
3. 对每个条目的 `pagelist` 做页码排序与区间整理
4. 计算条目所属分组
5. 转成输出层使用的 `IndexItem` 并追加到对应组

### 5.2 条目排序

`IndexEntrySlice.Less()` 会：

1. 逐层比较 `key`
2. 若 `key` 相同，再比较 `text`
3. 层级较短者优先

而单个字符串比较由 `Strcmp()` 承担，其排序语义是：

- 先判定字符串类型：空串、符号、数字开头混合串、纯数字、字母/汉字
- 类型不同直接按类型优先级比较
- 若两边都是纯十进制串，则按数值自然排序
- 否则逐 rune 比较，并把字符级比较委托给当前 collator
- 忽略大小写相等后，再以原始字符串稳定区分大小写

因此，排序框架负责通用比较顺序，具体中文字符如何比大小由 collator 策略决定。

### 5.3 页码整理

条目排序后，`PageSorter` 对每个 `IndexEntry.pagelist` 进一步处理：

1. 按严格或宽松规则对页码排序
2. 用栈匹配显式区间头尾
3. 生成 `[]PageRange`
4. 继续合并相邻单页和相邻区间

区间处理遵守两个主要开关：

- `strict`
  - 严格模式先按 `encap` 分离，再比较页码
  - 宽松模式优先按页码比较，`encap` 放到后面
- `disable_range`
  - 打开后停止把连续普通页自动并成区间，但仍会去重并保留显式区间

### 5.4 输出层数据结构

排序整理后的结构为：

#### `OutputIndex`

- `groups []IndexGroup`
- `style *OutputStyle`
- `option *OutputOptions`

#### `IndexGroup`

- `name string`
- `items []IndexItem`

#### `IndexItem`

- `level int`
- `text string`
- `page []PageRange`

#### `PageRange`

- `begin *Page`
- `end *Page`

这一步完成了一个重要降维：

- 输入层保留 `key`、完整多级路径和原始页码记录。
- 输出层只保留渲染需要的层级深度、显示文本和已经整理好的页码区间。

## 6. 输出层：`OutputIndex` 到 `.ind`

`OutputIndex.Output()` 是一层模板式渲染器，不再承担排序判断。

### 6.1 输出目标与编码

- 若未指定输出文件，则写到标准输出。
- 否则创建目标文件。
- 最终 writer 用 `transform.NewWriter(writer, option.encoder)` 包装，负责编码回目标字节流。

### 6.2 文本拼装顺序

输出顺序固定为：

1. `style.preamble`
2. 依次遍历各组
   - 组之间写 `group_skip`
   - 若启用标题，则写 `heading_prefix + group.name + heading_suffix`
3. 依次遍历组内条目
   - 按 `level` 选择 `item_0` / `item_1` / `item_2` 及其上下文相关变体
   - 输出 `item.text`
   - 调用 `writePage()` 输出页码区间
4. `style.postamble`

### 6.3 页码文本生成

`writePage()` 负责：

- 根据条目层级选择 `delim_0` / `delim_1` / `delim_2`
- 多个区间之间插入 `delim_n`
- 每个区间调用 `PageRange.Write()`
- 尾部补 `delim_t`

`PageRange.Write()` 再根据区间长度和样式参数选择：

- 单页输出
- 连续两页默认拆成两个独立页码，或在配置 `suffix_2p` 后用后缀简写
- 三页可用 `suffix_3p`
- 更长区间可用 `suffix_mp`
- 否则回退到 `begin + delim_r + end`
- 若区间带 `encap`，整体再包裹 `encap_prefix/infix/suffix`

## 7. 关键不变量与架构边界

### 7.1 稳定不变量

1. 所有语义处理都建立在 Unicode rune 上，编码只在边界处理。
2. 输入阶段一定先做去重和父级补全，再进入排序与输出阶段。
3. 排序层是唯一负责“条目顺序 + 页码区间语义”的地方。
4. 输出层只消费已经整理好的 `OutputIndex`，不应重新做业务判断。

### 7.2 当前实现边界

1. 输出显式只支持到三级索引；更深层级会记录日志并忽略。
2. `-p` 起始页码参数尚未实现。
3. `setpage_*`、`line_max`、`indent_*` 等样式字段当前解析后未形成完整输出行为。
4. 样式查找中的 kpathsea 集成目前只用于样式文件，不用于其它资源。
5. 页码类型判别较粗，按首字符决定页码格式。

## 8. 简化检索图

当需要沿数据流定位问题时，可按下列路径回看代码：

- 启动与编码：`main.go`
- 样式加载：`style.go`
- 输入读取与错误恢复：`input.go`、`numberedreader.go`
- 条目排序与页码归并：`sorter.go`、`pagenumber.go`
- 中文排序策略：`reading_collator.go`、`stroke_collator.go`、`radical_collator.go`
- 输出渲染：`output.go`
- 样式文件查找：`kpathsea/kpathsea.go`

把 zhmakeindex 视为“索引语言前端 + 中文排序中端 + 样式渲染后端”的三段式系统，通常最容易检索和定位问题。