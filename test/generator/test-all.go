/*
 * Copyright (C) 2020 ObjectBox Ltd. All rights reserved.
 * https://objectbox.io
 *
 * This file is part of ObjectBox Generator.
 *
 * ObjectBox Generator is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * ObjectBox Generator is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with ObjectBox Generator.  If not, see <http://www.gnu.org/licenses/>.
 */

package generator

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/objectbox/objectbox-generator/internal/generator"
	"github.com/objectbox/objectbox-generator/internal/generator/c"
	"github.com/objectbox/objectbox-generator/internal/generator/go"
	"github.com/objectbox/objectbox-generator/test/assert"
	"github.com/objectbox/objectbox-generator/test/build"
)

// this containing module name - used for test case modules
const moduleName = "github.com/objectbox/objectbox-go"

type lang struct {
	generatedExt string
	sourceExt    string
}

// codeLanguage returns a source code extension, currently everything is "go" except when it's in "testdata/c"
func codeLanguage(dir string) lang {
	if filepath.Base(dir) == "c" {
		return lang{generatedExt: "h", sourceExt: "fbs"}
	} else {
		return lang{generatedExt: "go", sourceExt: "go"}
	}
}

// generateAllDirs walks through the "data" and generates bindings for each subdirectory
// set overwriteExpected to TRUE to update all ".expected" files with the generated content
func generateAllDirs(t *testing.T, overwriteExpected bool) {
	var datadir = "testdata"
	folders, err := ioutil.ReadDir(datadir)
	assert.NoErr(t, err)

	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}

		var dir = filepath.Join(datadir, folder.Name())
		t.Run(folder.Name(), func(t *testing.T) {
			t.Parallel()
			generateOneDir(t, overwriteExpected, dir)
		})
	}
}

func generateOneDir(t *testing.T, overwriteExpected bool, srcDir string) {
	var dir = srcDir

	var errorTransformer = func(err error) error {
		return err
	}

	var cleanup = func() {}
	defer func() {
		cleanup()
	}()

	// Test in a temporary directory - if tested by an end user, the repo is read-only.
	// This doesn't apply if overwriteExpected is set, as that's only supposed to be run during this lib's development.
	if !overwriteExpected {
		tempRoot, err := ioutil.TempDir("", "objectbox-generator-test")
		assert.NoErr(t, err)

		// we can't defer directly because compilation step is run in a separate goroutine after this function exits
		cleanup = func() {
			assert.NoErr(t, os.RemoveAll(tempRoot))
		}

		// copy the source dir, including the relative paths (to make sure expected errors contain same paths)
		var tempDir = filepath.Join(tempRoot, srcDir)
		assert.NoErr(t, copyDirectory(srcDir, tempDir, 0700, 0600))
		t.Logf("Testing in a temporary directory %s", tempDir)

		// When outside of the project's directory, we need to set up the whole temp dir as its own module, otherwise
		// it won't find this `objectbox-go`. Therefore, we create a go.mod file pointing it to the right path.
		cwd, err := os.Getwd()
		assert.NoErr(t, err)
		var modulePath = "example.com/virtual/objectbox-generator/test/generator/" + srcDir
		var goMod = "module " + modulePath + "\n" +
			"replace " + moduleName + " => " + filepath.Join(cwd, "/../../") + "\n" +
			"require " + moduleName + " v0.0.0"
		assert.NoErr(t, ioutil.WriteFile(path.Join(tempDir, "go.mod"), []byte(goMod), 0600))

		// NOTE: we can't change directory using os.Chdir() because it applies to a process/thread, not a goroutine.
		// Therefore, we just map paths in received errors, so they match the expected ones.
		dir = tempDir
		errorTransformer = func(err error) error {
			if err == nil {
				return nil
			}
			var str = strings.Replace(err.Error(), tempRoot+string(os.PathSeparator), "", -1)
			str = strings.Replace(str, modulePath, moduleName+"/test/generator/"+srcDir, -1)
			return errors.New(str)
		}
	}

	modelInfoFile := generator.ModelInfoFile(dir)
	modelInfoExpectedFile := modelInfoFile + ".expected"

	modelFile := gogenerator.ModelFile(modelInfoFile)
	modelExpectedFile := modelFile + ".expected"

	// run the generation twice, first time with deleting old modelInfo
	for i := 0; i <= 1; i++ {
		if i == 0 {
			t.Logf("Testing %s without model info JSON", filepath.Base(dir))
			os.Remove(modelInfoFile)
		} else if testing.Short() {
			continue // don't test twice in "short" tests
		} else {
			t.Logf("Testing %s with previous model info JSON", filepath.Base(dir))
		}

		// setup the desired directory contents by copying "*.initial" files to their name without the extension
		initialFiles, err := filepath.Glob(filepath.Join(dir, "*.initial"))
		assert.NoErr(t, err)
		for _, initialFile := range initialFiles {
			assert.NoErr(t, copyFile(initialFile, initialFile[0:len(initialFile)-len(".initial")], 0))
		}

		generateAllFiles(t, overwriteExpected, dir, modelInfoFile, errorTransformer)

		assertSameFile(t, modelInfoFile, modelInfoExpectedFile, overwriteExpected)
		assertSameFile(t, modelFile, modelExpectedFile, overwriteExpected)
	}

	// verify the result can be built
	if !testing.Short() && codeLanguage(dir).generatedExt == "go" {
		// override the defer to prevent cleanup before compilation is actually run
		var cleanupAfterCompile = cleanup
		cleanup = func() {}

		t.Run("compile", func(t *testing.T) {
			defer cleanupAfterCompile()
			t.Parallel()

			var expectedError error
			if fileExists(path.Join(dir, "compile-error.expected")) {
				content, err := ioutil.ReadFile(path.Join(dir, "compile-error.expected"))
				assert.NoErr(t, err)
				expectedError = errors.New(string(content))
			}

			stdOut, stdErr, err := build.Package(dir)
			if err == nil && expectedError == nil {
				// successful
				return
			}

			if err == nil && expectedError != nil {
				assert.Failf(t, "Unexpected PASS during compilation")
			}

			// On Windows, we're getting a `go finding` message during the build - remove it to be consistent.
			var reg = regexp.MustCompile("go: finding " + moduleName + " v0.0.0[ \r\n]+")
			stdErr = reg.ReplaceAll(stdErr, nil)

			var receivedError = errorTransformer(fmt.Errorf("%s\n%s\n%s", stdOut, stdErr, err))

			// Fix paths in the error output on Windows so that it matches the expected error (which always uses '/').
			if os.PathSeparator != '/' {
				// Make sure the expected error doesn't contain the path separator already - to make it easier to debug.
				if strings.Contains(expectedError.Error(), string(os.PathSeparator)) {
					assert.Failf(t, "compile-error.expected contains this OS path separator '%v' so paths can't be normalized to '/'", string(os.PathSeparator))
				}
				receivedError = errors.New(strings.Replace(receivedError.Error(), string(os.PathSeparator), "/", -1))
			}

			assert.Eq(t, expectedError, receivedError)
		})
	}
}

func assertSameFile(t *testing.T, file string, expectedFile string, overwriteExpected bool) {
	// if no file is expected
	if !fileExists(expectedFile) {
		// there can be no source file either
		if fileExists(file) {
			assert.Failf(t, "%s is missing but %s exists", expectedFile, file)
		}
		return
	}

	content, err := ioutil.ReadFile(file)
	assert.NoErr(t, err)

	if overwriteExpected {
		assert.NoErr(t, copyFile(file, expectedFile, 0))
	}

	contentExpected, err := ioutil.ReadFile(expectedFile)
	assert.NoErr(t, err)

	if 0 != bytes.Compare(content, contentExpected) {
		assert.Failf(t, "generated file %s is not the same as %s", file, expectedFile)
	}
}

func generateAllFiles(t *testing.T, overwriteExpected bool, dir string, modelInfoFile string, errorTransformer func(error) error) {
	var modelFile = gogenerator.ModelFile(modelInfoFile)

	var lang = codeLanguage(dir) // go|c

	// remove generated files during development (they might be syntactically wrong)
	if overwriteExpected {
		files, err := filepath.Glob(filepath.Join(dir, "*.obx."+lang.generatedExt))
		assert.NoErr(t, err)

		for _, file := range files {
			assert.NoErr(t, os.Remove(file))
		}
	}

	// process all *.go files in the directory
	inputFiles, err := filepath.Glob(filepath.Join(dir, "*."+lang.sourceExt))
	assert.NoErr(t, err)
	for _, sourceFile := range inputFiles {
		// skip generated files & "expected results" files
		if strings.HasSuffix(sourceFile, ".obx."+lang.generatedExt) ||
			strings.HasSuffix(sourceFile, ".skip."+lang.sourceExt) ||
			strings.HasSuffix(sourceFile, "expected") ||
			strings.HasSuffix(sourceFile, "initial") ||
			sourceFile == modelFile {
			continue
		}

		t.Logf("  %s", filepath.Base(sourceFile))

		options := getOptions(t, lang, sourceFile, modelInfoFile)
		err = errorTransformer(generator.Process(sourceFile, options))

		// handle negative test
		var shouldFail = strings.HasSuffix(filepath.Base(sourceFile), ".fail."+lang.generatedExt)
		if shouldFail {
			if err == nil {
				assert.Failf(t, "Unexpected PASS on a negative test %s", sourceFile)
			} else {
				var errPlatformIndependent = strings.Replace(err.Error(), "\\", "/", -1)
				assert.Eq(t, getExpectedError(t, sourceFile).Error(), errPlatformIndependent)
				continue
			}
		}

		assert.NoErr(t, err)

		var bindingFile = gogenerator.BindingFile(sourceFile)
		var expectedFile = bindingFile + ".expected"
		assertSameFile(t, bindingFile, expectedFile, overwriteExpected)
	}
}

var generatorArgsRegexp = regexp.MustCompile("//go:generate go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen (.+)[\n|\r]")

func getOptions(t *testing.T, lang lang, sourceFile, modelInfoFile string) generator.Options {
	var options = generator.Options{
		ModelInfoFile: modelInfoFile,
		// NOTE zero seed for test-only - avoid changes caused by random numbers by fixing them to the same seed
		Rand:          rand.New(rand.NewSource(0)),
		CodeGenerator: &gogenerator.GoGenerator{},
	}

	if lang.generatedExt == "h" {
		options.CodeGenerator = &cgenerator.CGenerator{PlainC: true}
	}

	source, err := ioutil.ReadFile(sourceFile)
	assert.NoErr(t, err)

	var match = generatorArgsRegexp.FindSubmatch(source)
	if len(match) > 1 {
		var args = argsToMap(string(match[1]))

		setArgs(t, args, &options)
	}

	return options
}

var expectedErrorRegexp = regexp.MustCompile(`// *ERROR *=(.+)[\n|\r]`)
var expectedErrorRegexpMulti = regexp.MustCompile(`(?sU)/\* *ERROR.*[\n|\r](.+)\*/`)

func getExpectedError(t *testing.T, sourceFile string) error {
	source, err := ioutil.ReadFile(sourceFile)
	assert.NoErr(t, err)

	if match := expectedErrorRegexp.FindSubmatch(source); len(match) > 1 {
		return errors.New(strings.TrimSpace(string(match[1]))) // this is a "positive" return
	}

	if match := expectedErrorRegexpMulti.FindSubmatch(source); len(match) > 1 {
		return errors.New(strings.TrimSpace(string(match[1]))) // this is a "positive" return
	}

	assert.Failf(t, "missing error declaration in %s - add comment to the file // ERROR = expected error text", sourceFile)
	return nil
}

func setArgs(t *testing.T, args map[string]string, options *generator.Options) {
	for name, value := range args {
		_ = value // get rid of the compiler warning until we start using some options with values

		switch name {
		case "byValue":
			options.ByValue = true
		default:
			t.Fatalf("unknown option '%s'", name)
		}
	}
}

func argsToMap(args string) map[string]string {
	var result = map[string]string{}

	for _, arg := range strings.Split(strings.TrimSpace(args), "-") {
		arg = strings.TrimSpace(arg)

		if len(arg) == 0 {
			continue
		}

		var pair = strings.Split(arg, " ")
		if len(pair) == 1 {
			result[pair[0]] = ""
		} else {
			result[pair[0]] = pair[1]
		}
	}

	return result
}
