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
	"os"
	"path/filepath"

	"github.com/neuronlabs/strcase"
	"github.com/spf13/cobra"

	"github.com/neuronlabs/neuron/neuron-generator/internal/ast"
)

// collectionsCmd represents the models command
var collectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "Generates collection for the model's repository query access.",
	Long: `This generator allows to create collection for provided Model used by other neuron components.
The collection is a struct that allows to create and execute type safe queries for provided input model type.
A model type is provided with flag '-type' i.e.:

neuron-generator models collections -type MyModel -o ./collections`,
	Run: generateCollections,
}

func init() {
	modelsCmd.AddCommand(collectionsCmd)

	collectionsCmd.PersistentFlags().StringP("output", "o", "", "defines the output directory where the generated files should be located")
}

func generateCollections(cmd *cobra.Command, args []string) {
	typeNames, err := cmd.Flags().GetStringSlice("type")
	if err != nil {
		cmd.Usage()
		os.Exit(2)
	}

	g := ast.NewModelGenerator("snake", typeNames, tags)
	if len(args) == 0 {
		args = []string{"."}
	}

	// Parse provided argument packages.
	g.ParsePackages(args)

	// Extract all models from given packages.
	if err := g.ExtractPackageModels(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Extract the directory name from the arguments.
	dir := directory(args)

	// generate collection files.
	for _, collection := range g.Collections() {
		// Create new file if not exists.
		fileName := filepath.Join(dir, strcase.ToSnake(collection.Model.Name)+"_neuron_collection.go")
		modelFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Writing file: %s failed: %v\n", fileName, err)
			os.Exit(1)
		}
		if err = templates.ExecuteTemplate(modelFile, "collection", collection); err != nil {
			modelFile.Close()
			fmt.Fprintf(os.Stderr, "Error: execute model template failed: %v\n", err)
			os.Exit(1)
		}
		modelFile.Close()
	}
}
