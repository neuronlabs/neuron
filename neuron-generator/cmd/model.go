/*
Copyright © 2020 Jacek Kucharczyk kucjac@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/tools/imports"
)

// modelsCmd represents the models command
var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Generates model methods implementations, collections",
}

func init() {
	rootCmd.AddCommand(modelsCmd)

	modelsCmd.PersistentFlags().StringSliceP("type", "t", nil, "Specify the models type name (comma separated types)")
	modelsCmd.PersistentFlags().StringSliceP("exclude", "x", nil, "Define the types to exclude from the generator")
}

func directory(args []string) string {
	if len(args) == 1 && isDirectory(args[0]) {
		return args[0]
	}
	if len(tags) != 0 {
		log.Fatal("-tags option applies only to directories, not when files are specified")
	}
	return filepath.Dir(args[0])
}

func generateFile(fileName, templateName string, buf *bytes.Buffer, templateValue interface{}) {
	var err error
	if err = templates.ExecuteTemplate(buf, templateName, templateValue); err != nil {
		fmt.Fprintf(os.Stderr, "Error: execute model template failed: %v\n", err)
		os.Exit(1)
	}
	var result []byte
	switch codeFormatting {
	case gofmtFormat:
		result, err = format.Source(buf.Bytes())
	case goimportsFormat:
		err = os.Remove(fileName)
		if err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: deleting file failed: %v", err)
			os.Exit(1)
		}
		result, err = imports.Process(fileName, buf.Bytes(), nil)
	default:
		result = buf.Bytes()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: formatting go file failed: %v", err)
		os.Exit(1)
	}
	buf.Reset()
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Writing file: %s failed: %v\n", fileName, err)
		os.Exit(1)
	}
	defer file.Close()

	_, err = file.Write(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Writing file: '%s' failed. %v\n", fileName, err)
		os.Exit(1)
	}
}
