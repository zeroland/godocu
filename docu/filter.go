package docu

import (
	"go/ast"
	"go/doc"
	"os"
	"path/filepath"
	"strings"
)

// DefaultFilter 缺省的文件名过滤规则. 过滤掉非 ".go" 和 "_test.go" 结尾的文件
func DefaultFilter(name string) bool {
	return ShowTestFilter(name) && !strings.HasSuffix(name, "_test.go")
}

// ShowTestFilter 允许 "_test.go" 结尾的文件
func ShowTestFilter(name string) bool {
	return len(name) > 0 && name[0] != '_' && name[0] != '.' && strings.HasSuffix(name, ".go")
}

// ExportedFileFilter 剔除 non-nil file 中所有非导出声明, 返回该 file 是否具有导出声明.
func ExportedFileFilter(file *ast.File) bool {
	for i := 0; i < len(file.Decls); {
		if ExportedDeclFilter(file.Decls[i]) {
			i++
			continue
		}
		copy(file.Decls[i:], file.Decls[i+1:])
		file.Decls = file.Decls[:len(file.Decls)-1]
	}
	return len(file.Decls) != 0
}

// ExportedDeclFilter 剔除 non-nil decl 中所有非导出声明, 返回该 decl 是否具有导出声明.
func ExportedDeclFilter(decl ast.Decl) bool {
	switch decl := decl.(type) {
	case *ast.FuncDecl:
		if decl.Recv != nil && !exportedRecvFilter(decl.Recv) {
			return false
		}
		return decl.Name.IsExported()
	case *ast.GenDecl:
		for i := 0; i < len(decl.Specs); {
			if ExportedSpecFilter(decl.Specs[i]) {
				i++
				continue
			}
			copy(decl.Specs[i:], decl.Specs[i+1:])
			decl.Specs = decl.Specs[:len(decl.Specs)-1]
		}
		return len(decl.Specs) != 0
	}
	return false
}

// exportedRecvFilter 该方法仅仅适用于检测 ast.FuncDecl.Recv 是否导出
func exportedRecvFilter(fieldList *ast.FieldList) bool {

	for i := 0; i < len(fieldList.List); i++ {
		switch n := fieldList.List[i].Type.(type) {
		case *ast.Ident:
			if !n.IsExported() {
				return false
			}
		case *ast.StarExpr:
			ident, ok := n.X.(*ast.Ident)
			if !ok || !ident.IsExported() {
				return false
			}
		}
	}
	return true
}

// ExportedSpecFilter 剔除 non-nil spec 中所有非导出声明, 返回该 spec 是否具有导出声明.
func ExportedSpecFilter(spec ast.Spec) bool {
	switch n := spec.(type) {
	case *ast.ImportSpec:
		return true
	case *ast.ValueSpec:
		for i := 0; i < len(n.Names); {
			if n.Names[i].IsExported() {
				i++
				continue
			}
			copy(n.Names[i:], n.Names[i+1:])
			n.Names = n.Names[:len(n.Names)-1]
		}
		return len(n.Names) != 0
	case *ast.TypeSpec:
		if !n.Name.IsExported() {
			break
		}
		exportedExprFilter(n.Type)
	}
	return false
}

// 剔除非导出成员, 返回是否还有可见成员
func exportedExprFilter(expr ast.Expr) bool {
	switch n := expr.(type) {
	case *ast.StructType:
		if n.Fields == nil {
			break
		}
		list := n.Fields.List
		for i := 0; i < len(list); {
			names := list[i].Names
			for i := 0; i < len(names); {
				if names[i].IsExported() {
					i++
					continue
				}
				copy(names[i:], names[i+1:])
				names = names[:len(names)-1]
			}
			list[i].Names = names
			if len(names) != 0 {
				i++
				continue
			}
			copy(list[i:], list[i+1:])
			list = list[:len(list)-1]
		}
		n.Fields.List = list
		return len(list) != 0
	}

	return true
}

// WalkPath 可遍历 paths 及其子目录或者独立的文件.
// 若 paths 是包路径或绝对路径, 调用 walk 遍历 paths.
func WalkPath(paths string, walk filepath.WalkFunc) error {
	root := Abs(paths)
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return walk(root, info, err)
	}

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		ispkg := IsPkgDir(info)
		if err == nil && info.IsDir() {
			if info.Name() == "src" {
				return nil
			}
			if !ispkg {
				return filepath.SkipDir
			}
		}
		if err != nil || ispkg {
			return walk(path, info, err)
		}
		return nil
	})
}

func IsPkgDir(fi os.FileInfo) bool {
	name := fi.Name()
	return fi.IsDir() && len(name) > 0 &&
		name[0] != '_' && name[0] != '.' &&
		name != "testdata" && name != "vendor"
}

func License(file *ast.File) (lic string) {
	for _, comm := range file.Comments {
		lic = comm.Text()
		if doc.Synopsis(lic) == "" &&
			"copyright" == strings.ToLower(strings.SplitN(lic, " ", 2)[0]) {
			break
		}
		lic = ""
	}
	return
}
