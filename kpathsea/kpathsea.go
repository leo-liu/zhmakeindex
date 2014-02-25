// 提供 kpathsea 库的基本文件查找功能
package kpathsea

import (
	"os"
	"os/exec"
	"strings"
)

// 以 C 中对应于 enum kpse_file_format_type，见源代码 types.h
type FileFormatType int

// 对应于 TL2013 中的 kpathsea 6.11
const (
	GF_FORMAT FileFormatType = iota
	PK_FORMAT
	ANY_GLYPH_FORMAT // ``any'' meaning gf or pk
	TFM_FORMAT
	AFM_FORMAT
	BASE_FORMAT
	BIB_FORMAT
	BST_FORMAT
	CNF_FORMAT
	DB_FORMAT
	FMT_FORMAT
	FONTMAP_FORMAT
	MEM_FORMAT
	MF_FORMAT
	MFPOOL_FORMAT
	MFT_FORMAT
	MP_FORMAT
	MPPOOL_FORMAT
	MPSUPPORT_FORMAT
	OCP_FORMAT
	OFM_FORMAT
	OPL_FORMAT
	OTP_FORMAT
	OVF_FORMAT
	OVP_FORMAT
	PICT_FORMAT
	TEX_FORMAT
	TEXDOC_FORMAT
	TEXPOOL_FORMAT
	TEXSOURCE_FORMAT
	TEX_PS_HEADER_FORMAT
	TROFF_FONT_FORMAT
	TYPE1_FORMAT
	VF_FORMAT
	DVIPS_CONFIG_FORMAT
	IST_FORMAT
	TRUETYPE_FORMAT
	TYPE42_FORMAT
	WEB2C_FORMAT
	PROGRAM_TEXT_FORMAT
	PROGRAM_BINARY_FORMAT
	MISCFONTS_FORMAT
	WEB_FORMAT
	CWEB_FORMAT
	ENC_FORMAT
	CMAP_FORMAT
	SFD_FORMAT
	OPENTYPE_FORMAT
	PDFTEX_CONFIG_FORMAT
	LIG_FORMAT
	TEXMFSCRIPTS_FORMAT
	LUA_FORMAT
	FEA_FORMAT
	CID_FORMAT
	MLBIB_FORMAT
	MLBST_FORMAT
	CLUA_FORMAT
	RIS_FORMAT
	BLTXML_FORMAT
	LAST_FORMAT // one past last index
)

// 基本文件查询函数调用 kpsewhich 的最简单的备用实现。
// 在 C 中对应于
// extern KPSEDLL string kpathsea_find_file (kpathsea kpse, const_string name,
//    kpse_file_format_type format,  boolean must_exist);
// 见源代码 tex-file.h
func FindFile(name string, format FileFormatType, mustExist bool) string {
	// 动态库调用并不更优，暂不使用 findFile_dynamic(name, format, mustExist)
	return findFile_external(name)
}

// 基本文件查询函数调用 kpsewhich 的最简单的备用实现。
// 没有 format 与 mustExist 参数。
func findFile_external(name string) string {
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
