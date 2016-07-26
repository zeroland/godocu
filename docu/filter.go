package docu

import (
	"go/ast"
	"path/filepath"
	"strings"
)

// DefaultFilter 缺省的文件名过滤规则. 过滤掉非 ".go" 和 "_test.go" 结尾的文件
func DefaultFilter(name string) bool {
	return filepath.Ext(name) == ".go" && !strings.HasSuffix(name, "_test.go")
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
		return n.Name.IsExported()
	}
	return false
}
