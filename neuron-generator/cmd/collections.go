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

// collectionsCmd represents the models command
var collectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "Generates collection for the model's repository query access.",
	Long: `This generator allows to create collection for provided Model used by other neuron components.
The collection is a struct that allows to create and execute type safe queries for provided input model type.
A model type is provided with flag '-type' i.e.:

neuron-generator models collections -type MyModel -o ./collections`,
	PreRun: modelsPreRun,
	Run:    generateCollections,
}

func init() {
	modelsCmd.AddCommand(collectionsCmd)

	// Here you will define your flags and configuration settings.
	collectionsCmd.Flags().StringP("naming-convention", "n", "snake", `set the naming convention for the output models. 
Possible values: 'snake', 'kebab', 'lower_camel', 'camel'`)
}

func generateCollections(cmd *cobra.Command, args []string) {
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
	buf := &bytes.Buffer{}
	if !g.HasCollectionInitializer() {
		fileName := filepath.Join(dir, "initialize_collections.neuron.go")
		generateFile(fileName, "initialize_collections", buf, g.CollectionInitializer(false))
	}

	var modelNames []string
	for _, collection := range g.Collections("") {
		// Create new file if not exists.
		fileName := filepath.Join(dir, strcase.ToSnake(collection.Model.Name)+"_collection.neuron.go")
		generateFile(fileName, "collection", buf, collection)
		modelNames = append(modelNames, collection.Model.Name)
	}
	fmt.Fprintf(os.Stdout, "Success. Generated collections for: %s models.\n", strings.Join(modelNames, ","))
}
