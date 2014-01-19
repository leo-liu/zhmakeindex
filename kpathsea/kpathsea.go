// 提供 kpathsea 库的基本查询功能
package kpathsea

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// 基本路径查询函数
// 在 C 中对应于
// extern KPSEDLL string kpathsea_path_search
//  (kpathsea kpse, const_string path, const_string name, boolean must_exist);
// 这里给出最简单的实现，must_exist 参数被忽略
func PathSearch(path, name string, must_exist bool) string {
	outpath := filepath.Join(path, name)
	// 先尝试直接搜索（速度较快）
	if _, err := os.Stat(outpath); err == nil {
		return outpath
	}
	// 调用 kpsewhich 外部程序搜索（慢）
	var cmd *exec.Cmd
	if path != "" {
		cmd = exec.Command("kpsewhich", "-path="+path, name)
	} else {
		cmd = exec.Command("kpsewhich", name)
	}
	out, err := cmd.Output()
	if err != nil {
		return ""
	} else {
		outpath = strings.TrimSpace(string(out))
		return outpath
	}
}
