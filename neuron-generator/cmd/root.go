// +build !codeanalysis

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
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/neuronlabs/strcase"
	"github.com/spf13/cobra"

	"github.com/neuronlabs/neuron/neuron-generator/bintemplates"
)

var (
	codeFormatting codeFormat
	tags           []string
	templates      *template.Template
)

// rootCmd represents the base command when called without any sub commands
var rootCmd = &cobra.Command{
	Use:   "neuron-generator",
	Short: "A code generator for the neuron package.",
	Long:  `It is a code generator for the Golang github.com/neuronlabs/neuron package.`,
}

func init() {
	rootCmd.PersistentFlags().StringSlice("tags", []string{}, "comma-separated list of build tags to apply")
	rootCmd.PersistentFlags().StringP("format", "f", "", "optional formatting of the output files. Possible values: gofmt, goimports")

	parseTemplates()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type codeFormat int

const (
	noFormat codeFormat = iota
	gofmtFormat
	goimportsFormat
)

func modelsPreRun(cmd *cobra.Command, args []string) {
	var err error
	tags, err = cmd.Flags().GetStringSlice("tags")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: flag - %v\n", err)
		cmd.Usage()
		os.Exit(2)
	}

	var codeFmt string
	codeFmt, err = cmd.Flags().GetString("format")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		cmd.Usage()
		os.Exit(2)
	}

	switch codeFmt {
	case "gofmt":
		codeFormatting = gofmtFormat
	case "goimports":
		codeFormatting = goimportsFormat
	default:
		codeFormatting = noFormat
	}
}

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

func parseTemplates() {
	functionMap := template.FuncMap{
		"toLowerCamel": strcase.ToLowerCamel,
		"toSnake":      strcase.ToSnake,
		"timestamp":    func() string { return time.Now().Format(time.RFC1123Z) },
	}
	templates = template.New("")
	for _, tmpl := range bintemplates.AssetNames() {
		_, err := templates.New("").Funcs(functionMap).Parse(string(bintemplates.MustAsset(tmpl)))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
