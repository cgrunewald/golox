package main

import (
	"bytes"
	"os"
	"path"
	"testing"
)

// TestGeneratedFilesUpToDate ensures the checked-in AST files match the
// generator output, i.e. that `go generate ./interpreter` has been run.
func TestGeneratedFilesUpToDate(t *testing.T) {
	for _, ast := range asts {
		expected, err := generate(ast.mainClass, ast.grammar)
		if err != nil {
			t.Fatalf("generate(%s) failed: %v", ast.mainClass, err)
		}

		actual, err := os.ReadFile(path.Join("..", "..", "interpreter", ast.fileName))
		if err != nil {
			t.Fatalf("could not read %s: %v", ast.fileName, err)
		}

		if !bytes.Equal(expected, actual) {
			t.Errorf("%s is out of date; run `go generate ./interpreter`", ast.fileName)
		}
	}
}

func TestGenerateIsFormatted(t *testing.T) {
	source, err := generate("Expr", []string{"Literal : Value interface{}"})
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	if !bytes.Contains(source, []byte("\tValue interface{}\n")) {
		t.Errorf("expected gofmt'd output with tab indentation, got:\n%s", source)
	}
}
