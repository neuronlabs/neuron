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
	"os"
	"path/filepath"
	"strings"

	"github.com/neuronlabs/strcase"
	"github.com/spf13/cobra"

	"github.com/neuronlabs/neuron/neuron-generator/internal/ast"
)

// modelMethodsCmd represents the methods command
var modelMethodsCmd = &cobra.Command{
	Use:   "methods",
	Short: "Generates neuron model basic field and relation methods.",
	Long: `This generator allows to create model interfaces used by other neuron components.
By default it creates github.com/neuronlabs/neuron/mapping model interfaces implementation 
for provided input model type. A model type is provided with flag '-type' i.e.:

neuron-generator model methods -type=MyModel
Model methods must exists in the same namespace package. Due to the fact that the generator 
creates these files in the same directory as input. 
By default generator takes current working directory as an input.`,
	PreRun: modelsPreRun,
	Run:    generateModelMethods,
}

func init() {
	modelsCmd.AddCommand(modelMethodsCmd)

	// Here you will define your flags and configuration settings.
	modelMethodsCmd.Flags().StringP("naming-convention", "n", "snake", `set the naming convention for the output models. 
Possible values: 'snake', 'kebab', 'lower_camel', 'camel'`)
}

func generateModelMethods(cmd *cobra.Command, args []string) {
	namingConvention, err := cmd.Flags().GetString("naming-convention")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		cmd.Usage()
		os.Exit(2)
	}

	switch namingConvention {
	case "kebab", "snake", "lower_camel", "camel":
	default:
		fmt.Fprintf(os.Stderr, "Error: provided unsupported naming convention: '%v'", namingConvention)
		cmd.Usage()
		os.Exit(2)
	}
	// Get the optional type names flag.
	typeNames, err := cmd.Flags().GetStringSlice("type")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading flags failed: '%v\n", err)
		os.Exit(2)
	}
	g := ast.NewModelGenerator(namingConvention, typeNames, tags)

	// Parse provided argument packages.
	g.ParsePackages([]string{"."})

	// Extract all models from given packages.
	if err := g.ExtractPackageModels(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Get the directory from the arguments.
	dir := directory(args)

	buf := &bytes.Buffer{}

	// Generate model files.
	var modelNames []string
	for _, model := range g.Models() {
		fileName := filepath.Join(dir, strcase.ToSnake(model.Name)+"_methods.neuron.go")
		generateFile(fileName, "model", buf, model)
		modelNames = append(modelNames, model.Name)
	}
	fmt.Fprintf(os.Stdout, "Success. Generated methods for: %s models.\n", strings.Join(modelNames, ","))
}
