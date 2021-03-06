package docu

import (
	"os"
	"path/filepath"
	"runtime"
)

var (
	GOROOT  = strOr(os.Getenv("GOROOT"), runtime.GOROOT())
	GOPATHS = filepath.SplitList(os.Getenv("GOPATH"))
)

// via go/build/syslist.go

const goosList = "android darwin dragonfly freebsd linux nacl netbsd openbsd plan9 solaris windows "
const goarchList = "386 amd64 amd64p32 arm armbe arm64 arm64be ppc64 ppc64le mips mipsle mips64 mips64le mips64p32 mips64p32le ppc s390 s390x sparc sparc64 "

// Warehouse 为预定义托管仓库域名.
// 因托管商差异, 依照 Part 计算的仓库地址不一定正确.
var Warehouse = []struct {
	Host string // 域名
	Part int    // 仓库路径占用的段数
}{
	{"github.com", 2},
	{"gopkg.in", 1},
	{"bitbucket.org", 2},
	{"code.google.com", 2},
	{"golang.org", 2},
	{"google.golang.org", 1},
	{"launchpad.net", 2},
	{"git.oschina.net", 2},
}

func strOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func existsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func existsFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// Abs 返回 path 的绝对路径.
// 如果 path 疑似绝对路径返回 path.
// 否则在 GOROOT, GOPATHS 中搜索 path 并返回绝对路径.
// 如果未找到返回 path.
func Abs(path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	if path[0] == '.' {
		if abs, err := filepath.Abs(path); err == nil && exists(abs) {
			return abs
		}
	}

	abs, err := filepath.Abs(filepath.Join(GOROOT, "src", path))
	if err == nil && exists(abs) {
		return abs
	}

	for _, gopath := range GOPATHS {
		abs, err := filepath.Abs(filepath.Join(gopath, "src", path))
		if err == nil && exists(abs) {
			return abs
		}
	}

	return path
}
