# 样式文件参考

## 文档目的

本文汇总 zhmakeindex 支持的 `.ist/.mst` 样式关键字、默认值与兼容性边界。它是样式相关任务的稳定查阅表，而不是教程。

样式系统的总体规则如下：

- 程序总是先加载内置默认值，再用样式文件中的键覆盖。
- 样式文件名若无扩展名，会自动补 `.ist`。
- 若未显式给出 `-s`，且只有一个输入文件，则会优先尝试同名 `.mst`。
- 样式文件路径先按当前路径检查，找不到时再通过 `kpsewhich` 搜索。
- 未识别关键字只记日志并忽略，不会报错终止。

## 1. 样式文件语法

### 1.1 token 形式

`ScanStyleTokens` 支持以下 token：

- 普通标识符/数字
- 单引号字符，如 `'{'`
- 双引号字符串，如 `"\n  \\item "`
- 反引号字符串，如 `` `原样文本` ``

### 1.2 注释与空白

- `%` 到行尾为注释
- 空白字符会被跳过

注意：样式文件自身的注释起始符固定为 `%`，不受输入样式关键字 `comment` 影响。

### 1.3 取值类型约定

- rune 型字段通常写成单引号字符
- 字符串型字段通常写成双引号或反引号字符串
- 整数型字段写为十进制数字

## 2. 输入样式关键字

这些关键字决定 `.idx` 解析语法，映射到 `InputStyle`。

| 关键字 | 默认值 | 含义 | 兼容性说明 |
| --- | --- | --- | --- |
| `keyword` | `"\\indexentry"` | 索引记录命令名 | 与 makeindex 默认输入语法兼容 |
| `arg_open` | `'{'` | 参数起始符 | 兼容 |
| `arg_close` | `'}'` | 参数结束符 | 兼容 |
| `actual` | `'@'` | 排序键与显示文本分隔符 | 兼容 |
| `encap` | `'|'` | encapsulation 命令起始符 | 兼容 |
| `escape` | `'\\'` | 转义符 | 兼容 |
| `level` | `'!'` | 多级索引层级分隔符 | 兼容 |
| `quote` | `'"'` | 引号保护符 | 兼容 |
| `page_compositor` | `"-"` | 复合页码分隔符 | makeindex 兼容；zhmakeindex 允许多字符字符串 |
| `range_open` | `'('` | 显式页码区间开始标记 | 兼容 |
| `range_close` | `')'` | 显式页码区间结束标记 | 兼容 |
| `comment` | `'%'` | `.idx` 输入中的行注释起始符 | zhmakeindex 扩展，不是传统 makeindex 样式核心项 |

### 2.1 输入样式补充说明

- `page_compositor` 在解析页码时通过字符串拆分实现，因此不局限于单字符。
- `comment` 只影响 `.idx` 输入扫描，不影响 `.ist` 文件本身的注释语法。
- 输入解析支持引号、转义和嵌套花括号深度跟踪，因此即使更改语法字符，也仍遵循同一状态机结构。

## 3. 输出样式关键字

这些关键字映射到 `OutputStyle`，控制 `.ind` 渲染、分组标题与页码展示。

| 关键字 | 默认值 | 含义 | 兼容性说明 |
| --- | --- | --- | --- |
| `preamble` | `"\\begin{theindex}\n"` | 输出前导文本 | 兼容 |
| `postamble` | `"\n\n\\end{theindex}\n"` | 输出尾部文本 | 兼容 |
| `setpage_prefix` | `"\n  \\setcounter{page}{"` | 起始页码前缀 | 保留兼容字段；当前未见完整生效路径 |
| `setpage_suffix` | `"}\n"` | 起始页码后缀 | 保留兼容字段；当前未见完整生效路径 |
| `group_skip` | `"\n\n  \\indexspace\n"` | 组间分隔文本 | 兼容 |
| `headings_flag` | `0` | 分组标题开关 | 兼容 |
| `lethead_flag` | 无单独默认值 | `headings_flag` 兼容别名 | 历史兼容别名 |
| `heading_prefix` | `""` | 分组标题前缀 | 兼容 |
| `lethead_prefix` | 无单独默认值 | `heading_prefix` 兼容别名 | 历史兼容别名 |
| `heading_suffix` | `""` | 分组标题后缀 | 兼容 |
| `lethead_suffix` | 无单独默认值 | `heading_suffix` 兼容别名 | 历史兼容别名 |
| `symhead_positive` | `"Symbols"` | 正标题模式下的符号组标题 | 兼容 |
| `symhead_negative` | `"symbols"` | 负标题模式下的符号组标题 | 兼容 |
| `numhead_positive` | `"Numbers"` | 正标题模式下的数字组标题 | 兼容 |
| `numhead_negative` | `"numbers"` | 负标题模式下的数字组标题 | 兼容 |
| `stroke_prefix` | `""` | 笔画组标题前缀 | zhmakeindex 中文扩展 |
| `stroke_suffix` | `" 画"` | 笔画组标题后缀 | zhmakeindex 中文扩展 |
| `radical_prefix` | `""` | 部首组标题前缀 | zhmakeindex 中文扩展 |
| `radical_suffix` | `"部"` | 部首组标题后缀 | zhmakeindex 中文扩展 |
| `radical_simplify_flag` | `1` | 是否显示简化部首提示 | 解析关键字使用该名称；与字段命名存在历史不一致 |
| `radical_simplified_prefix` | `"（"` | 简化部首提示前缀 | zhmakeindex 中文扩展 |
| `radical_simplified_suffix` | `"）"` | 简化部首提示后缀 | zhmakeindex 中文扩展 |
| `item_0` | `"\n  \\item "` | 一级条目前缀 | 兼容 |
| `item_1` | `"\n    \\subitem "` | 二级条目前缀 | 兼容 |
| `item_2` | `"\n      \\subsubitem "` | 三级条目前缀 | 兼容 |
| `item_01` | `"\n    \\subitem "` | 一级后直接进入二级时的特殊前缀 | 兼容 |
| `item_x1` | `"\n    \\subitem "` | 父级无页码时进入二级的特殊前缀 | 兼容 |
| `item_12` | `"\n      \\subsubitem "` | 二级后直接进入三级时的特殊前缀 | 兼容 |
| `item_x2` | `"\n      \\subsubitem "` | 父级无页码时进入三级的特殊前缀 | 兼容 |
| `delim_0` | `", "` | 一级条目和页码之间分隔符 | 兼容 |
| `delim_1` | `", "` | 二级条目和页码之间分隔符 | 兼容 |
| `delim_2` | `", "` | 三级条目和页码之间分隔符 | 兼容 |
| `delim_n` | `", "` | 多个页码或区间之间分隔符 | 兼容 |
| `delim_r` | `"--"` | 区间连接符 | 兼容 |
| `delim_t` | `""` | 页码列表尾随文本 | 兼容 |
| `encap_prefix` | `"\\"` | encap 输出前缀 | 兼容 |
| `encap_infix` | `"{"` | encap 名称与页码之间的连接文本 | 兼容 |
| `encap_suffix` | `"}"` | encap 输出后缀 | 兼容 |
| `page_precedence` | `"rnaRA"` | 页码类型优先级串 | 兼容 |
| `line_max` | `72` | 折行最大长度 | 保留兼容字段；当前未见完整生效路径 |
| `indent_space` | `"\t\t"` | 折行缩进文本 | 保留兼容字段；当前未见完整生效路径 |
| `indent_length` | `16` | 折行缩进长度 | 保留兼容字段；当前未见完整生效路径 |
| `suffix_2p` | `""` | 两页区间简写后缀 | 当前代码已生效 |
| `suffix_3p` | `""` | 三页区间简写后缀 | 当前代码已生效 |
| `suffix_mp` | `""` | 多页区间简写后缀 | 当前代码已生效 |

## 4. 分组标题相关语义

### 4.1 `headings_flag`

- `0`：不输出分组标题
- `> 0`：输出“正向”标题，例如 `Symbols`、`Numbers`、`A`..`Z`
- `< 0`：输出“负向/小写”标题，例如 `symbols`、`numbers`、`a`..`z`

### 4.2 中文扩展分组

下列字段只在中文扩展排序方式中使用：

- 笔画排序：`stroke_prefix`、`stroke_suffix`
- 部首排序：`radical_prefix`、`radical_suffix`、`radical_simplify_flag`、`radical_simplified_prefix`、`radical_simplified_suffix`

拼音排序不会使用这些字段，因为其汉字并入 A..Z 组。

## 5. 页码相关关键字解释

### 5.1 `page_precedence`

默认值 `rnaRA` 表示页码类型优先级为：

1. 小写罗马数字 `r`
2. 阿拉伯数字 `n`
3. 小写字母 `a`
4. 大写罗马数字 `R`
5. 大写字母 `A`

若该串中出现未知字符，程序会记录日志并回退到默认映射。

### 5.2 `suffix_2p` / `suffix_3p` / `suffix_mp`

`PageRange.Write()` 当前已经使用这三项：

- 两页区间且设置了 `suffix_2p` → 输出 `起始页 + suffix_2p`
- 三页区间且设置了 `suffix_3p` → 输出 `起始页 + suffix_3p`
- 更长区间且设置了 `suffix_mp` → 输出 `起始页 + suffix_mp`
- 否则退回 `begin + delim_r + end`

若连续两页是由普通单页自动合并而来，且 `suffix_2p` 为空，则会继续输出成两个独立页码，而不是范围表达式。

### 5.3 encap 包裹

当页码区间带 `encap` 时，最终文本形式为：

- `encap_prefix + encap + encap_infix + rangestr + encap_suffix`

默认值即典型的 `\command{...}` 形式。

## 6. 当前实现边界与兼容说明

### 6.1 保留但未完全生效的字段

根据当前代码路径，以下字段虽然能解析并保存，但未看到完整输出行为：

- `setpage_prefix`
- `setpage_suffix`
- `line_max`
- `indent_space`
- `indent_length`

它们应视为“兼容保留项”，而不是当前稳定可依赖的输出能力。

### 6.2 已知命名不一致

存在一个需要特别记住的兼容点：

- 结构体字段名是 `radical_simplified_flag`
- 样式解析关键字接受的是 `radical_simplify_flag`

因此稳定文档应以“解析关键字名”为准，也就是样式文件里要写 `radical_simplify_flag`。

### 6.3 输出层级上限

样式字段虽然定义了 `item_0` 到 `item_2` 及其变体，但当前输出逻辑只显式支持三级索引：

- level 0 → `item_0`
- level 1 → `item_1` / `item_01` / `item_x1`
- level 2 → `item_2` / `item_12` / `item_x2`
- 更深层级 → 记录日志并忽略

因此，样式系统当前也只对三级输出结构稳定生效。

## 7. makeindex 兼容结论

从稳定契约看，可以把兼容性总结为：

1. `.ist` 基本语法与 makeindex 兼容。
2. 传统输入关键字和大多数输出关键字兼容。
3. 对未知关键字采用宽容忽略策略。
4. 增加了中文排序分组相关关键字。
5. 增加了反引号字符串与 `.idx` 注释字符等扩展。
6. 某些传统字段仅保留解析兼容，当前实现未完全兑现对应排版能力。

## 8. 检索提示

样式相关问题可按下面路径定位：

- 样式解析与默认值：`style.go`
- 输入语法如何消费这些字段：`input.go`
- 页码字段如何影响输出：`output.go`、`pagenumber.go`
- 样式查找：`kpathsea/kpathsea.go`

若问题表现为“某个样式关键字为什么不生效”，先确认它是输入语法字段、输出模板字段，还是仅为兼容保留的未完全实现字段。