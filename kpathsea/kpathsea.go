// 提供 kpathsea 库的基本文件查找功能
package kpathsea // import "github.com/leo-liu/zhmakeindex/kpathsea"

import (
	"os"
	"os/exec"
	"strings"
)

// 基本文件查询函数，调用 kpsewhich 实现。
// 在 C 中对应于
// extern KPSEDLL string kpathsea_find_file (kpathsea kpse, const_string name,
//    kpse_file_format_type format,  boolean must_exist);
// 但这里后几个参数不使用。
func FindFile(name string) string {
	// 先尝试直接搜索（速度较快）
	if _, err := os.Stat(name); err == nil {
		return name
	}
	// 调用 kpsewhich 外部程序搜索（慢）
	cmd := exec.Command("kpsewhich", name)
	out, err := cmd.Output()
	if err != nil {
		return ""
	} else {
		outpath := strings.TrimSpace(string(out))
		return outpath
	}
}
