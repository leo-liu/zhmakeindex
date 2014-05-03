// +build windows,386

package kpathsea

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

// TeX Live 2014 使用的库
var kpse_dll_name string = "kpathsea620.dll"

var dll *syscall.LazyDLL
var kpse uintptr

func init() {
	var err error

	dll = syscall.NewLazyDLL(kpse_dll_name)
	if err = dll.Load(); err != nil {
		kpse = 0
		fmt.Println("Error:", err)
		return
	}

	kpse_new := dll.NewProc("kpathsea_new")
	kpse, _, err = kpse_new.Call()
	if errno := err.(syscall.Errno); errno != 0 {
		kpse = 0
		fmt.Println("Error:", errno)
		return
	}

	kpse_set_program_name := dll.NewProc("kpathsea_set_program_name")
	program_name_ptr, _ := syscall.BytePtrFromString(os.Args[0])
	_, _, err = kpse_set_program_name.Call(kpse, uintptr(unsafe.Pointer(program_name_ptr)))
	if errno := err.(syscall.Errno); errno != 0 {
		kpse = 0
		fmt.Println("Error:", errno)
		return
	}

	runtime.SetFinalizer(dll, func(_ *syscall.LazyDLL) {
		fmt.Println("Finished")
		kpse_finish := dll.NewProc("kpathsea_finish")
		kpse_finish.Call(kpse)
	})
}

func findFile_dynamic(name string, format FileFormatType, mustExist bool) string {
	var mustExist_int uintptr = 0
	if mustExist {
		mustExist_int = 1
	}
	// 初始化失败
	if kpse == 0 {
		return ""
	}
	// 调用 kpathsea_find_file 找文件
	kpse_find_file := dll.NewProc("kpathsea_find_file")
	name_ptr, _ := syscall.BytePtrFromString(name)
	path_uintptr, _, _ := kpse_find_file.Call(kpse, uintptr(unsafe.Pointer(name_ptr)), uintptr(format), mustExist_int)
	// 找不到
	if path_uintptr == 0 {
		return ""
	}
	// 找到了
	path_ptr := (*byte)(unsafe.Pointer(path_uintptr))
	var path_bytes []byte
	for *path_ptr != 0 { // 逐字节附加返回的串
		path_bytes = append(path_bytes, *path_ptr)
		path_ptr = (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(path_ptr)) + 1))
	}
	return string(path_bytes)
}
