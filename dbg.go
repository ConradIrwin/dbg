// package dbg provides a more helpful [fmt.Println] for debugging in Go.
//
// dbg.Dbg wraps [spew.Dump] to additionally print the name of each argument and the
// filename and line number of the caller.
package dbg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

// Dbg pretty-prints its arguments with their names, types and values to os.Stdout.
func Dbg(args ...any) {
	impl(os.Stdout, args...)
}

// To pretty-prints its arguments with their names, types and values to the given Writer.
func To(w io.Writer, args ...any) {
	impl(w, args...)
}

func impl(w io.Writer, args ...any) {
	_, file, line, _ := runtime.Caller(2)
	argNames := extractFunctionCallArgs(file, line)
	fileName := filepath.Base(file)
	fmt.Fprintf(w, "%s:%d:", fileName, line)
	if len(args) == 0 {
		fmt.Fprintln(w)
		return
	}
	if len(argNames) != len(args) {
		spew.Fdump(w, args...)
		return
	}
	for i, arg := range args {
		dumped := strings.TrimSpace(spew.Sdump(arg))
		fmt.Fprintf(w, " %s = %s", argNames[i], strings.ReplaceAll(dumped, "\n", "  \n"))
		if strings.Contains(dumped, "\n") || i == len(args)-1 {
			fmt.Fprint(w, "\n")
		} else {
			fmt.Fprint(w, ",")
		}
	}
}

func extractFunctionCallArgs(filename string, lineNum int) (args []string) {
	sourceBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return nil
	}

	for n := range ast.Preorder(file) {
		if fset.Position(n.Pos()).Line != lineNum {
			continue
		}
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			continue
		}
		if fn, ok := callExpr.Fun.(*ast.SelectorExpr); !ok {
			continue
		} else if fn.Sel.Name != "Dbg" && fn.Sel.Name != "To" {
			continue
		}

		for _, arg := range callExpr.Args {
			start := fset.Position(arg.Pos()).Offset
			end := fset.Position(arg.End()).Offset
			if start >= 0 && end <= len(sourceBytes) {
				args = append(args, string(sourceBytes[start:end]))
			}
		}
		break
	}

	return args
}
