//go:build windows
// +build windows

package cli

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

// addToSystemPath 添加目录到 Windows 环境变量 PATH
func addToSystemPath(dir string) error {
	advapi32 := syscall.NewLazyDLL("advapi32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")
	regOpenKeyEx := advapi32.NewProc("RegOpenKeyExW")
	regQueryValueEx := advapi32.NewProc("RegQueryValueExW")
	regSetValueEx := advapi32.NewProc("RegSetValueExW")
	regCloseKey := advapi32.NewProc("RegCloseKey")
	sendMessageTimeout := user32.NewProc("SendMessageTimeoutW")

	const (
		HKEY_CURRENT_USER = 0x80000001
		KEY_READ          = 0x20019
		KEY_WRITE         = 0x20006
		REG_SZ            = 1
		HWND_BROADCAST    = 0xFFFF
		WM_SETTINGCHANGE  = 0x001A
		SMTO_ABORTIFHUNG  = 0x0002
	)

	keyPath, _ := syscall.UTF16PtrFromString(`Environment`)
	valueName, _ := syscall.UTF16PtrFromString(`Path`)

	var hKey syscall.Handle
	ret, _, _ := regOpenKeyEx.Call(
		uintptr(HKEY_CURRENT_USER),
		uintptr(unsafe.Pointer(keyPath)),
		0,
		KEY_READ|KEY_WRITE,
		uintptr(unsafe.Pointer(&hKey)),
	)
	if ret != 0 {
		return fmt.Errorf("打开注册表失败: %d", ret)
	}
	defer regCloseKey.Call(uintptr(hKey))

	var bufSize uint32 = 4096
	buf := make([]uint16, bufSize)
	ret, _, _ = regQueryValueEx.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valueName)),
		0,
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bufSize)),
	)
	if ret != 0 && ret != 234 {
		return fmt.Errorf("读取 PATH 失败: %d", ret)
	}

	currentPath := syscall.UTF16ToString(buf)
	paths := strings.Split(currentPath, ";")
	for _, p := range paths {
		if strings.EqualFold(strings.TrimSpace(p), dir) {
			return nil // 已存在
		}
	}

	newPath := currentPath
	if !strings.HasSuffix(currentPath, ";") && currentPath != "" {
		newPath += ";"
	}
	newPath += dir

	newPathUTF16, _ := syscall.UTF16FromString(newPath)
	ret, _, _ = regSetValueEx.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valueName)),
		0,
		REG_SZ,
		uintptr(unsafe.Pointer(&newPathUTF16[0])),
		uintptr(len(newPathUTF16)*2),
	)
	if ret != 0 {
		return fmt.Errorf("写入 PATH 失败: %d", ret)
	}

	envPtr, _ := syscall.UTF16PtrFromString("Environment")
	sendMessageTimeout.Call(
		HWND_BROADCAST,
		WM_SETTINGCHANGE,
		0,
		uintptr(unsafe.Pointer(envPtr)),
		SMTO_ABORTIFHUNG,
		5000,
		0,
	)

	return nil
}

// removeFromSystemPath 从 Windows 环境变量 PATH 中移除目录
func removeFromSystemPath(dir string) error {
	advapi32 := syscall.NewLazyDLL("advapi32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")
	regOpenKeyEx := advapi32.NewProc("RegOpenKeyExW")
	regQueryValueEx := advapi32.NewProc("RegQueryValueExW")
	regSetValueEx := advapi32.NewProc("RegSetValueExW")
	regCloseKey := advapi32.NewProc("RegCloseKey")
	sendMessageTimeout := user32.NewProc("SendMessageTimeoutW")

	const (
		HKEY_CURRENT_USER = 0x80000001
		KEY_READ          = 0x20019
		KEY_WRITE         = 0x20006
		REG_SZ            = 1
		HWND_BROADCAST    = 0xFFFF
		WM_SETTINGCHANGE  = 0x001A
		SMTO_ABORTIFHUNG  = 0x0002
	)

	keyPath, _ := syscall.UTF16PtrFromString(`Environment`)
	valueName, _ := syscall.UTF16PtrFromString(`Path`)

	var hKey syscall.Handle
	ret, _, _ := regOpenKeyEx.Call(
		uintptr(HKEY_CURRENT_USER),
		uintptr(unsafe.Pointer(keyPath)),
		0,
		KEY_READ|KEY_WRITE,
		uintptr(unsafe.Pointer(&hKey)),
	)
	if ret != 0 {
		return fmt.Errorf("打开注册表失败: %d", ret)
	}
	defer regCloseKey.Call(uintptr(hKey))

	var bufSize uint32 = 4096
	buf := make([]uint16, bufSize)
	ret, _, _ = regQueryValueEx.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valueName)),
		0,
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bufSize)),
	)
	if ret != 0 && ret != 234 {
		return fmt.Errorf("读取 PATH 失败: %d", ret)
	}

	currentPath := syscall.UTF16ToString(buf)
	paths := strings.Split(currentPath, ";")
	found := false
	newPaths := make([]string, 0)
	for _, p := range paths {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" && !strings.EqualFold(trimmed, dir) {
			newPaths = append(newPaths, trimmed)
		} else if strings.EqualFold(trimmed, dir) {
			found = true
		}
	}

	if !found {
		return nil // 不存在
	}

	newPath := strings.Join(newPaths, ";")
	newPathUTF16, _ := syscall.UTF16FromString(newPath)
	ret, _, _ = regSetValueEx.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valueName)),
		0,
		REG_SZ,
		uintptr(unsafe.Pointer(&newPathUTF16[0])),
		uintptr(len(newPathUTF16)*2),
	)
	if ret != 0 {
		return fmt.Errorf("写入 PATH 失败: %d", ret)
	}

	envPtr, _ := syscall.UTF16PtrFromString("Environment")
	sendMessageTimeout.Call(
		HWND_BROADCAST,
		WM_SETTINGCHANGE,
		0,
		uintptr(unsafe.Pointer(envPtr)),
		SMTO_ABORTIFHUNG,
		5000,
		0,
	)

	return nil
}