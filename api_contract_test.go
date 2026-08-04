package agentx_test

import (
	"context"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"reflect"
	"sort"
	"strings"
	"testing"

	agentx "github.com/wsnacj/agentx-go"
)

type exactClientContract interface {
	Run(context.Context, agentx.RunRequest) (agentx.RunResult, error)
	Shutdown(context.Context) error
}

var (
	_ exactClientContract                         = (*agentx.Client)(nil)
	_ func(agentx.Config) (*agentx.Client, error) = agentx.New
)

func TestExactPackageExports(t *testing.T) {
	expected := []string{
		"AdapterRunRequest",
		"AdapterRunResult",
		"Client",
		"CodeCanceled",
		"CodeClientClosed",
		"CodeDeadlineExceeded",
		"CodeExecutionFailed",
		"CodeInvalidArgument",
		"CodeShutdownFailed",
		"CodeUnsupportedProfile",
		"Config",
		"DefaultExecutionProfile",
		"Error",
		"ErrorCode",
		"ExecutionAdapter",
		"ExecutionProfile",
		"New",
		"RunRequest",
		"RunResult",
	}

	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, ".", func(info fs.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	parsed := packages["agentx"]
	files := make([]*ast.File, 0, len(parsed.Files))
	for _, file := range parsed.Files {
		files = append(files, file)
	}
	checked, err := (&types.Config{Importer: importer.Default()}).Check(
		"github.com/wsnacj/agentx-go",
		fset,
		files,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	var actual []string
	for _, name := range checked.Scope().Names() {
		if token.IsExported(name) {
			actual = append(actual, name)
		}
	}
	sort.Strings(expected)
	sort.Strings(actual)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("exported package identifiers = %v, want %v", actual, expected)
	}
}

func TestExactErrorCodeValues(t *testing.T) {
	expected := map[agentx.ErrorCode]string{
		agentx.CodeInvalidArgument:    "invalid_argument",
		agentx.CodeCanceled:           "canceled",
		agentx.CodeDeadlineExceeded:   "deadline_exceeded",
		agentx.CodeClientClosed:       "client_closed",
		agentx.CodeUnsupportedProfile: "unsupported_profile",
		agentx.CodeExecutionFailed:    "execution_failed",
		agentx.CodeShutdownFailed:     "shutdown_failed",
	}
	if len(expected) != 7 {
		t.Fatalf("ErrorCode value count = %d, want 7", len(expected))
	}
	for code, value := range expected {
		if string(code) != value {
			t.Errorf("ErrorCode %q = %q, want %q", code, string(code), value)
		}
	}
}

func TestExactStructFieldsAndMethods(t *testing.T) {
	assertExportedFields(t, reflect.TypeOf(agentx.AdapterRunRequest{}), []string{
		"Input string",
		"SessionID string",
	})
	assertExportedFields(t, reflect.TypeOf(agentx.AdapterRunResult{}), []string{
		"RunID string",
		"SessionID string",
		"Status string",
		"Reply string",
	})
	assertExportedFields(t, reflect.TypeOf(agentx.ExecutionProfile{}), []string{
		"Activation string",
		"ControlMode string",
		"ExecutionIntensity string",
		"Driver string",
		"ResultPolicy string",
		"Lifecycle string",
	})
	assertExportedFields(t, reflect.TypeOf(agentx.Config{}), []string{
		"Adapter agentx.ExecutionAdapter",
		"Profile agentx.ExecutionProfile",
	})
	assertExportedFields(t, reflect.TypeOf(agentx.RunRequest{}), []string{
		"Input string",
		"SessionID string",
	})
	assertExportedFields(t, reflect.TypeOf(agentx.RunResult{}), []string{
		"RunID string",
		"SessionID string",
		"Status string",
		"Reply string",
		"Evidence []string",
		"Blockers []string",
		"NextAction string",
		"Profile agentx.ExecutionProfile",
	})
	assertExportedFields(t, reflect.TypeOf(agentx.Error{}), []string{
		"Code agentx.ErrorCode",
		"Retryable bool",
		"Message string",
	})

	clientType := reflect.TypeOf((*agentx.Client)(nil))
	if clientType.NumMethod() != 2 {
		t.Fatalf("*Client exported methods = %d, want 2", clientType.NumMethod())
	}
	for _, name := range []string{"Run", "Shutdown"} {
		if _, ok := clientType.MethodByName(name); !ok {
			t.Errorf("*Client is missing %s", name)
		}
	}

	errorType := reflect.TypeOf((*agentx.Error)(nil))
	if errorType.NumMethod() != 3 {
		t.Fatalf("*Error exported methods = %d, want 3", errorType.NumMethod())
	}
	for _, name := range []string{"Error", "Is", "Unwrap"} {
		if _, ok := errorType.MethodByName(name); !ok {
			t.Errorf("*Error is missing %s", name)
		}
	}

	adapterType := reflect.TypeOf((*agentx.ExecutionAdapter)(nil)).Elem()
	if adapterType.NumMethod() != 3 {
		t.Fatalf("ExecutionAdapter methods = %d, want 3", adapterType.NumMethod())
	}
	for _, name := range []string{"ClassifyError", "Run", "Shutdown"} {
		if _, ok := adapterType.MethodByName(name); !ok {
			t.Errorf("ExecutionAdapter is missing %s", name)
		}
	}
}

func assertExportedFields(t *testing.T, typ reflect.Type, expected []string) {
	t.Helper()
	var actual []string
	for index := 0; index < typ.NumField(); index++ {
		field := typ.Field(index)
		if field.PkgPath == "" {
			actual = append(actual, field.Name+" "+field.Type.String())
		}
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%s exported fields = %v, want %v", typ, actual, expected)
	}
}
