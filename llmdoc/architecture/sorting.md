# 中文排序系统

## 文档目的

本文说明 zhmakeindex 的中文排序子系统如何在统一框架下支持拼音、笔画和部首三种策略，并解释运行期比较算法与构建期 CJK 数据表之间的关系。

这个子系统的核心目标不是“只给汉字排个序”，而是让包含符号、数字、拉丁字母和汉字的索引项能够在 makeindex 风格工作流中产生可预期的中文分组和稳定排序结果。

## 1. 总体设计：统一排序框架 + 可替换 collator

### 1.1 `IndexCollator` 接口

`sorter.go` 用 `IndexCollator` 抽象出所有与排序策略相关的变化点：

- `InitGroups(style *OutputStyle) []IndexGroup`
  - 初始化分组名称与分组数量
- `Group(entry *IndexEntry) int`
  - 根据条目首字符决定放入哪个组
- `RuneCmp(a, b rune) int`
  - 比较两个字符的大小
- `IsLetter(r rune) bool`
  - 判断某字符是否应归入“字母/汉字类”

这个接口把“通用排序框架”和“具体中文规则”分离开了。

### 1.2 `IndexSorter` 的职责

`IndexSorter` 只是一个持有 `IndexCollator` 的壳层，负责把命令行参数映射到具体策略：

- `pinyin` / `reading` → `ReadingIndexCollator`
- `bihua` / `stroke` → `StrokeIndexCollator`
- `bushou` / `radical` → `RadicalIndexCollator`

策略选定后，`SortIndex()` 负责统一流程：

1. 初始化分组
2. 对输入条目整体排序
3. 对每个条目的页码列表排序并归并
4. 计算每个条目的分组
5. 组装成 `OutputIndex`

因此，排序系统的变化点只应放在 collator 中；排序流水线本身不因具体中文策略而分叉。

## 2. 统一字符串比较算法

### 2.1 `Strcmp` 的层次

索引项比较最终会下沉到 `IndexEntrySlice.Strcmp(a, b string)`。它不是简单字典序，而是一个分层比较器：

1. 先调用 `getStringType()` 把字符串分成几类：
   - 空串 `EMPTY_STR`
   - 符号开头 `SYMBOL_STR`
   - 数字开头但不是纯数字 `NUM_SYMBOL_STR`
   - 纯数字 `NUM_STR`
   - 字母或汉字 `LETTER_STR`
2. 类型不同，直接按类型优先级比较。
3. 如果两边都是纯十进制字符串，优先做自然数比较，而不是纯文本比较。
4. 否则把字符串拆为 rune 序列，逐字符比较，并把单字符比较委托给当前 collator 的 `RuneCmp`。
5. 如果忽略大小写后的比较仍相等，再回退到原始字符串比较，以获得稳定、可重复的排序顺序。

### 2.2 多级索引比较

`IndexEntrySlice.Less()` 的比较顺序为：

1. 逐层比较 `level[i].key`
2. 若 `key` 相同，再比较 `level[i].text`
3. 所有已比较层级都相同，则层级更短的条目排前

这意味着：

- 排序首先尊重隐藏排序键 `key`
- 显示文本 `text` 只在 key 相同时作为细化比较条件
- 父级条目天然排在同前缀的子条目前

### 2.3 `RuneCmpIgnoreCases`

当某个字符不属于当前 collator 的 CJK 数据覆盖范围时，三种 collator 都会回退到 `RuneCmpIgnoreCases()`：

- 先对两个 rune 做 `unicode.ToLower`
- 再按码点比较

这保证了英文、符号和未收录字符仍有统一、稳定的排序结果。

## 3. 三种排序方式的实现差异

三种 collator 都遵循同一骨架：

- 符号组固定在最前
- 数字组固定在其后
- 拉丁字母 A..Z 组放在数字之后
- 中文条目根据具体策略继续映射到不同的扩展组

不同点在于“汉字如何映射到组”和“字符如何比较”。

### 3.1 拼音排序：`ReadingIndexCollator`

#### 分组

拼音排序只生成：

- 符号组
- 数字组
- A..Z 26 个字母组

如果条目首字符是汉字，且在 `CJK.Readings` 中有读音，则会取其标准化拼音串的首字母，把该条目并入对应的字母组。

#### 比较规则

`RuneCmp(a, b)` 的优先级是：

1. 两边都没有拼音数据 → 回退到忽略大小写码点比较
2. 只有一边有拼音数据 → 有拼音数据的一边视作“字母/汉字类”，排在无拼音数据字符之后
3. 两边都有拼音数据 → 先比较标准化拼音串
4. 拼音串相同 → 用原始 Unicode 码点打破平局

#### 适用语义

这是最接近现代中文读者习惯的索引排序方式，也是默认排序方式。

### 3.2 笔画排序：`StrokeIndexCollator`

#### 分组

笔画排序生成：

- 符号组
- 数字组
- A..Z 组
- 从 1 到 `MAX_STROKE` 的笔画数组

若汉字在 `CJK.Strokes` 中存在条目，则按笔顺串长度确定组号。组标题由 `stroke_prefix + 笔画数 + stroke_suffix` 形成。

#### 比较规则

字符比较顺序为：

1. 两边都没有笔顺数据 → 回退到忽略大小写码点比较
2. 只有一边有笔顺数据 → 有笔顺数据的一边作为汉字类参与排序
3. 总笔画数不同 → 笔画少者在前
4. 总笔画数相同 → 比较完整笔顺编码串
5. 笔顺编码相同 → 用 Unicode 码点打破平局

#### 适用语义

适用于不依赖读音、强调字形检索的索引或字表场景。

### 3.3 部首排序：`RadicalIndexCollator`

#### 分组

部首排序生成：

- 符号组
- 数字组
- A..Z 组
- 214 个康熙部首组

如果条目首字符在 `CJK.RadicalStrokes` 中有记录，则组号由其部首号决定。组名来自 `CJK.Radicals`，并受以下样式参数影响：

- `radical_prefix`
- `radical_suffix`
- `radical_simplified_flag`
- `radical_simplified_prefix`
- `radical_simplified_suffix`

当某部首有简化写法且样式允许显示时，标题会以“正体部首 + 简化提示”的形式生成。

#### 比较规则

部首排序复用了一个非常紧凑的数据编码：`RadicalStroke`。

每个值是一个字符串，其字节布局为：

1. 第一字节：部首号
2. 第二字节：余画数
3. 后续字节：字符自身 UTF-8 编码

由于这个编码可直接参与字符串比较，因此 `RuneCmp` 只需要：

1. 若两边都无部首数据，回退到忽略大小写码点比较
2. 若只有一边有部首数据，按有/无数据区分
3. 若两边都有部首数据，直接比较编码串
4. 相同则再以 Unicode 码点打破平局

#### 适用语义

适用于传统字典式索引和部首检索习惯较强的场景。

## 4. CJK 运行期数据结构

排序系统的性能依赖于运行期直接查表，而不是在比较时临时推导汉字元数据。

### 4.1 `CJK.Readings`

- 类型：`map[rune]string`
- 值：标准化后的拼音串，如 `ling2`
- 用途：拼音排序的分组与字符比较

### 4.2 `CJK.Strokes`

- 类型：`map[rune]string`
- 值：笔顺编码串
- 语义：字符串长度就是总笔画数；字符串内容可进一步表达完整笔顺顺序
- 用途：笔画排序的分组与细粒度比较

### 4.3 `CJK.RadicalStrokes`

- 类型：`map[rune]RadicalStroke`
- 值：部首号 + 余画 + 字符自身的紧凑编码串
- 用途：部首排序的分组与直接可比较的排序键

### 4.4 `CJK.Radicals`

- 类型：`[MAX_RADICAL + 1]Radical`
- 字段：`Origin`、`Simplified`
- 用途：构造 214 部首的显示名称

## 5. 构建期数据生成机制

运行期使用的 CJK 表不是手工维护，而是由 `CJK/maketables.go` 生成。该程序带有 `// +build ignore` 标记，说明它是工具程序，不参与正常构建。

### 5.1 数据来源

- 拼音数据：`Unihan_Readings.txt`
- 笔画数据：`sunwb_strokeorder.txt` 与 `Unihan_DictionaryLikeData.txt`
- 部首数据：`CJKRadicals.txt` 与 `Unihan_IRGSources.txt`
- 上游总包：在线下载的 `Unihan.zip`

### 5.2 生成思路

#### 拼音表

- 提取 `kMandarin` 与 `kHanyuPinyin`
- 做标准化，转成带声调数字的拼音
- 生成 `readings.go`

#### 笔顺表

- 优先使用本地笔顺数据得到完整笔顺序列
- 再用 Unihan 的总笔画数字段补全缺失字符
- 对只有笔画数没有笔顺的字符，用固定占位笔画编码补足长度
- 生成 `strokes.go`

#### 部首表

- 读取 214 个部首的正体/简体信息
- 读取每个汉字的部首号与余画数
- 编码成可直接比较的 `RadicalStroke` 字符串
- 生成 `radicalstrokes.go`

### 5.3 Unicode 覆盖与工程特征

- `MAX_CODEPOINT = 0x40000`，覆盖 Unicode 第 0 到第 3 平面
- 数据更新流程可复现，但依赖在线下载 `latest` Unihan 数据，因此版本锁定性一般
- 当前仓库中的已生成表头显示 Unicode 版本为 10.0.0

## 6. 特殊处理与边界

### 6.1 “〇”的特殊处理

系统明确把 `〇` 视为汉字而非普通数字：

- `IsNumRune()` 中排除了 `〇`
- 拼音表会补入 `ling2`
- 部首表与笔画表也会补相应数据

这保证了它在中文索引中能进入合适的汉字排序逻辑，而不是落入数字组。

### 6.2 多音字限制

拼音排序依赖静态的“常用读音”表，不做上下文消歧。因此：

- 一个汉字通常只选一个代表性读音
- 同音字最终仍需靠码点打破平局
- 这是数据驱动策略带来的稳定性与语义精细度之间的取舍

### 6.3 未收录字符的回退行为

如果字符不在对应 CJK 数据表中：

- 不会导致排序失败
- 会回退到忽略大小写的 Unicode 码点比较
- 分组也可能回退到符号组，而不是中文扩展组

因此，数据表质量直接影响中文索引排序体验，但不会破坏程序整体可运行性。

## 7. 与页码排序的关系

中文排序系统只决定：

- 条目如何比较
- 条目落在哪个分组

页码如何排序、区间如何合并并不属于 `IndexCollator` 的职责，而是由 `PageSorter` 统一处理。也就是说：

- 条目排序与页码排序是两个平行子系统
- 两者只在 `SortIndex()` 中汇合
- 更换拼音/笔画/部首策略不会改变页码处理规则

这种解耦是整个排序架构能保持稳定的原因之一。

## 8. 检索提示

当任务涉及排序问题时，可按下面路径定位：

- 排序总流程与 `Strcmp`：`sorter.go`
- 拼音策略：`reading_collator.go`
- 笔画策略：`stroke_collator.go`
- 部首策略：`radical_collator.go`
- 运行期数据表：`CJK/readings.go`、`CJK/strokes.go`、`CJK/radicalstrokes.go`
- 数据生成器：`CJK/maketables.go`

若问题表现为“为什么某汉字落到这个组”“为什么两个条目这样排序”，先确认所选 collator，再检查对应数据表是否存在该字符记录，通常最有效。