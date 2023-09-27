package main

import (
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// THIS SCRIPT IS EXPECTED TO RUN IN THE ROOT DIRECTORY OF THE PROJECT

const (
	cloudFunctionsDir = "./cloudfunction/"
)

func main() {
	functionTarget, ok := os.LookupEnv("FUNCTION_TARGET")
	if !ok {
		panic("FUNCTION_TARGET env var must be set")
	}
	cloudFunctionFileName := functionTarget + ".go"
	cloudFunctionFullPath := cloudFunctionsDir + cloudFunctionFileName

	// Create temp dir if it does not exist
	if _, err := os.Stat("./tmp"); os.IsNotExist(err) {
		if err := os.Mkdir("./tmp", 0777); err != nil {
			panic(err)
		}
	}

	// Create FUNCTION_TARGET-specific temp dir
	tmpDir, err := os.MkdirTemp("./tmp", functionTarget)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy cloud function file into temp dir
	if _, err := exec.Command("cp", cloudFunctionFullPath, filepath.Join(tmpDir, cloudFunctionFileName)).Output(); err != nil {
		panic(string(err.(*exec.ExitError).Stderr))
	}

	// Execute `go list -m` to get the module path
	stdout, err := exec.Command("go", "list", "-m").Output()
	if err != nil {
		panic(string(err.(*exec.ExitError).Stderr))
	}
	module := strings.TrimSpace(string(stdout))

	// Recursively parse local imports of cloud function file
	imps := resolveLocalImportsFile(make(map[string]struct{}), cloudFunctionFullPath, module)

	// Copy local packages into the temp dir
	for _, imp := range imps {
		// Create directory for local package
		tdDir := filepath.Join(tmpDir, imp)
		if err := os.MkdirAll(tdDir, 0777); err != nil {
			panic(err)
		}

		if _, err := exec.Command("cp", "-r", imp, filepath.Dir(tdDir)).Output(); err != nil {
			panic(string(err.(*exec.ExitError).Stderr))
		}
	}

	// Remove test/testdata folder if it exists
	if err := os.RemoveAll(filepath.Join(tmpDir, "test", "testdata")); err != nil {
		panic(err)
	}

	// Copy go.mod file into temp dir
	if _, err := exec.Command("cp", "go.mod", filepath.Join(tmpDir, "go.mod")).Output(); err != nil {
		panic(string(err.(*exec.ExitError).Stderr))
	}

	// Execute `go mod tidy` in temp dir to remove unused deps
	goModTidy := exec.Command("go", "mod", "tidy")
	goModTidy.Dir = tmpDir
	if _, err := goModTidy.Output(); err != nil {
		panic(string(err.(*exec.ExitError).Stderr))
	}

	// Zip files
	var zipFile = functionTarget + ".zip"
	files, err := filepath.Glob(filepath.Join(tmpDir, "*"))
	if err != nil {
		panic(err)
	}
	var args []string = []string{"-r", zipFile}
	for _, file := range files {
		_, after, _ := strings.Cut(file, "/")   // Remove ./tmp from path
		_, after2, _ := strings.Cut(after, "/") // Remove temp dir from path
		args = append(args, after2)
	}
	zip := exec.Command("zip", args...)
	zip.Dir = tmpDir
	if stdout, err := zip.Output(); err != nil {
		panic(string(stdout))
	}

	// Copy zip file out
	copyZip := exec.Command("cp", zipFile, "../..")
	copyZip.Dir = tmpDir
	if _, err := copyZip.Output(); err != nil {
		panic(string(err.(*exec.ExitError).Stderr))
	}
}

func resolveLocalImportsFile(set map[string]struct{}, path string, module string) []string {
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		panic(err)
	}

	for _, imp := range f.Imports {
		depModule := strings.Trim(imp.Path.Value, "\"")

		if !strings.HasPrefix(depModule, module) { // Is not local import
			continue
		}

		_, after, _ := strings.Cut(depModule, module)
		depRelPath := strings.Trim(after, "/")
		if _, ok := set[depRelPath]; ok {
			continue
		}
		set[depRelPath] = struct{}{}

		for _, imp := range resolveLocalImportsDir(set, depRelPath, module) {
			if _, ok := set[imp]; ok {
				continue
			}
			set[imp] = struct{}{}
		}
	}

	var imps []string = make([]string, 0, len(set))
	for imp := range set {
		imps = append(imps, imp)
	}

	return imps
}

func resolveLocalImportsDir(set map[string]struct{}, path string, module string) []string {
	pkgs, err := parser.ParseDir(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		for fileName := range pkg.Files {
			for _, imp := range resolveLocalImportsFile(set, fileName, module) {
				if _, ok := set[imp]; ok {
					continue
				}
				set[imp] = struct{}{}
			}
		}
	}

	var imps []string = make([]string, 0, len(set))
	for imp := range set {
		imps = append(imps, imp)
	}

	return imps
}
