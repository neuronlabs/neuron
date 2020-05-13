package ast

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/neuronlabs/inflection"
	"github.com/neuronlabs/strcase"
	"golang.org/x/tools/go/packages"

	"github.com/neuronlabs/neuron/neuron-generator/input"
)

const (
	kindInt     = "int"
	kindInt8    = "int8"
	kindInt16   = "int16"
	kindInt32   = "int32"
	kindInt64   = "int64"
	kindUint    = "uint"
	kindUint8   = "uint8"
	kindUint16  = "uint16"
	kindUint32  = "uint32"
	kindUint64  = "uint64"
	kindFloat32 = "float32"
	kindFloat64 = "float64"
	kindByte    = "byte"
	kindRune    = "rune"
	kindString  = "string"
	kindBool    = "bool"
	kindUintptr = "uintptr"
	kindPointer = "Pointer"
	kindNil     = "nil"
)

// NewModelGenerator creates new model generator.
func NewModelGenerator(namingConvention string, types, tags []string) *ModelGenerator {
	gen := &ModelGenerator{
		models:              map[string]*input.Model{},
		modelsFiles:         map[*input.Model]string{},
		Types:               types,
		Tags:                tags,
		importFields:        map[string]map[string][]*ast.Ident{},
		imports:             map[string]string{},
		modelImportedFields: map[*input.Model][]*importField{},
	}
	switch namingConvention {
	case "kebab":
		gen.namerFunc = strcase.ToKebab
	case "lower_camel":
		gen.namerFunc = strcase.ToLowerCamel
	case "camel":
		gen.namerFunc = strcase.ToCamel
	default:
		gen.namerFunc = strcase.ToSnake
	}
	return gen
}

// ModelGenerator is the neuron model generator.
type ModelGenerator struct {
	namerFunc           func(s string) string
	pkgs                []*packages.Package
	Tags                []string
	Types               []string
	imports             map[string]string
	importFields        map[string]map[string][]*ast.Ident
	models              map[string]*input.Model
	modelsFiles         map[*input.Model]string
	modelImportedFields map[*input.Model][]*importField
}

// Collections return generator collections.
func (g *ModelGenerator) Collections(packageName string) (collections []*input.CollectionInput) {
	for _, model := range g.Models() {
		collections = append(collections, model.CollectionInput(packageName))
	}
	return collections
}

// HasCollectionInitializer checks if the package contains collection initializer.
func (g *ModelGenerator) HasCollectionInitializer() bool {
	var rootPkg string
	for _, model := range g.Models() {
		rootPkg = model.PackageName
	}

	for _, pkg := range g.pkgs {
		if rootPkg != pkg.Name {
			continue
		}
		for _, file := range pkg.CompiledGoFiles {
			fmt.Printf("File: %s\n", file)
			if file == "initialize_collections.neuron.go" {
				return true
			}
		}
	}
	return false
}

// CollectionInitializer gets collection initializer.
func (g *ModelGenerator) CollectionInitializer(externalController bool) *input.Collections {
	model := g.Models()[0]
	col := &input.Collections{PackageName: model.PackageName, ExternalController: externalController}
	if externalController {
		col.Imports.Add("github.com/neuronlabs/neuron/controller")
	}
	return col
}

// ExtractPackageModels extracts all models for provided in the packages.
func (g *ModelGenerator) ExtractPackageModels() error {
	// Find all struct Types that might be potential models
	for _, pkg := range g.pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				d, ok := decl.(*ast.GenDecl)
				if ok && d.Tok == token.TYPE {
					for _, model := range g.extractFileModels(d, file) {
						g.modelsFiles[model] = filepath.Join(pkg.PkgPath, file.Name.Name)
						model.PackageName = pkg.Name
					}
				}
			}
		}
	}

	for _, modelName := range g.Types {
		if _, ok := g.models[modelName]; !ok {
			return fmt.Errorf("model: '%s' not found", modelName)
		}
	}

	if len(g.models) == 0 {
		return errors.New("no models found")
	}

	if err := g.parseImportPackages(); err != nil {
		return err
	}

	// Find the most common receiver for the models.
	for _, pkg := range g.pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				dt, ok := decl.(*ast.FuncDecl)
				if ok && dt.Recv != nil {
					g.getMethodReceivers(dt)
				}
			}
		}
	}

	// Search and set most common receiver for each model.
	for _, model := range g.models {
		g.setReceiver(model)
	}
	// Sort model fields.
	for _, model := range g.models {
		model.SortFields()
	}
	return nil
}

// Models returns generator models.
func (g *ModelGenerator) Models() (models []*input.Model) {
	for model := range g.modelsFiles {
		models = append(models, model)
	}
	return models
}

// ParsePackages analyzes the single package constructed from the patterns and Tags.
// ParsePackages exits if there is an error.
func (g *ModelGenerator) ParsePackages(patterns []string) {
	cfg := &packages.Config{
		Mode:  packages.NeedSyntax | packages.NeedImports | packages.NeedDeps | packages.NeedFiles | packages.NeedName,
		Tests: true,
	}
	if len(g.Tags) > 0 {
		cfg.BuildFlags = []string{fmt.Sprintf("-Tags=%s", strings.Join(g.Tags, " "))}
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}
	if len(pkgs) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No packages found:\n")
		os.Exit(1)
	}
	g.pkgs = pkgs
}

func (g *ModelGenerator) setReceiver(model *input.Model) {
	// Get the most common receiver from existing methods for given model.
	var (
		mostCommonReceiver string
		maxCount           int
	)
	for name, count := range model.Receivers {
		if maxCount > count {
			maxCount = count
			mostCommonReceiver = name
		}
	}
	// If no receivers were found yet - set the receiver to the lowered first letter of the model name.
	if mostCommonReceiver == "" {
		mostCommonReceiver = strings.ToLower(model.Name[:1])
	}
	model.Receiver = mostCommonReceiver
}

func (g *ModelGenerator) getMethodReceivers(dt *ast.FuncDecl) {
	for _, r := range dt.Recv.List {
		if len(r.Names) == 0 {
			continue
		}
		switch t := r.Type.(type) {
		case *ast.StarExpr:
			ident, ok := t.X.(*ast.Ident)
			if !ok {
				continue
			}
			model, ok := g.models[ident.Name]
			if ok {
				model.Receivers[r.Names[0].Name]++
			}
		case *ast.Ident:
			model, ok := g.models[t.Name]
			if ok {
				model.Receivers[r.Names[0].Name]++
			}
		}
	}
}

func (g *ModelGenerator) extractFileModels(d *ast.GenDecl, file *ast.File) (models []*input.Model) {
	for _, spec := range d.Specs {
		switch st := spec.(type) {
		case *ast.TypeSpec:
			structType, ok := st.Type.(*ast.StructType)
			if !ok {
				continue
			}
			if st.Name == nil {
				continue
			}
			modelName := st.Name.Name
			if len(g.Types) != 0 {
				var matchedModel string
				for i, tp := range g.Types {
					if tp == st.Name.Name {
						matchedModel = tp
						g.Types = append(g.Types[:i], g.Types[i+1:]...)
						break
					}
				}
				if matchedModel == "" {
					continue
				}
			}
			if model := g.extractModel(file, structType, modelName); model != nil {
				models = append(models, model)
			}
		default:
			continue
		}
	}
	return models
}

func (g *ModelGenerator) extractModel(file *ast.File, structType *ast.StructType, modelName string) (model *input.Model) {
	model = &input.Model{
		CollectionName: g.namerFunc(inflection.Plural(modelName)),
		Name:           modelName,
		Receivers:      make(map[string]int),
	}

	// Find primary field key.
	for i, structField := range structType.Fields.List {
		if len(structField.Names) == 0 {
			// Embedded fields are not taken into account.
			continue
		}
		name := structField.Names[0]
		if !name.IsExported() {
			// Private fields are not taken into account.
			continue
		}

		if !isExported(structField) {
			continue
		}
		field := input.Field{Index: i, Name: name.String(), NeuronName: g.namerFunc(name.String()), Type: fieldTypeName(structField.Type), Model: model}

		// Set the Tags for given field.
		if structField.Tag != nil {
			field.Tags = structField.Tag.Value
			tags := extractTags(field.Tags, "neuron", ";", ",")
			for _, tag := range tags {
				if tag.key == "-" {
					continue
				}
			}
		}

		if isFieldRelation(structField) {
			field.IsSlice = isMany(structField.Type)
			field.IsElemPointer = isElemPointer(structField)
			field.IsPointer = isPointer(structField)
			model.Relations = append(model.Relations, &field)
			continue
		} else if importedField := g.isImported(file, structField); importedField != nil {
			importedField.Field = &field
			importedField.AstField = structField
			if isPrimary(structField) {
				model.Primary = importedField.Field
			}
			g.modelImportedFields[model] = append(g.modelImportedFields[model], importedField)
			continue
		}
		fieldPtr := &field
		g.setModelField(structField, fieldPtr, false)
		// Check if field is a primary key field.
		if isPrimary(structField) {
			model.Primary = fieldPtr
		} else {
			model.Fields = append(model.Fields, fieldPtr)
		}
	}

	if model.Primary == nil {
		return nil
	}
	defaultModelPackages := []string{
		"github.com/neuronlabs/neuron/errors",
		"github.com/neuronlabs/neuron/mapping",
	}
	for _, pkg := range defaultModelPackages {
		model.Imports.Add(pkg)
	}

	g.models[model.Name] = model

	for _, relation := range model.Relations {
		if relation.IsSlice {
			model.MultiRelationer = true
		} else {
			model.SingleRelationer = true
		}
	}
	if len(model.Fields) > 0 {
		model.Fielder = true
	}

	for _, importedField := range g.modelImportedFields[model] {
		g.imports[importedField.Path] = importedField.Ident.Name
		pkgTypes := g.importFields[importedField.Path]
		if pkgTypes == nil {
			pkgTypes = map[string][]*ast.Ident{}
		}
		pkgTypes[importedField.Ident.Name] = append(pkgTypes[importedField.Ident.Name], importedField.Ident)
		g.importFields[importedField.Path] = pkgTypes
	}
	return model
}

func isElemPointer(field *ast.Field) bool {
	return isExprElemPointer(field.Type)
}

func isExprElemPointer(expr ast.Expr) bool {
	switch x := expr.(type) {
	case *ast.StarExpr:
		return isExprElemPointer(x.X)
	case *ast.ArrayType:
		_, isStar := x.Elt.(*ast.StarExpr)
		return isStar
	}
	return false
}

func (g *ModelGenerator) parseImportPackages() error {
	if len(g.imports) == 0 {
		return nil
	}
	var pkgPaths []string
	for pkg := range g.importFields {
		pkgPaths = append(pkgPaths, pkg)
	}
	cfg := &packages.Config{
		Mode:       packages.NeedSyntax,
		BuildFlags: g.Tags,
	}
	pkgs, err := packages.Load(cfg, pkgPaths...)
	if err != nil {
		return err
	}
	if packages.PrintErrors(pkgs) > 1 {
		return errors.New("error while loading import packages")
	}
	for _, pkg := range pkgs {
		importTypes := g.importFields[pkg.ID]
		cnt := len(importTypes)
	fileLoop:
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				if cnt == 0 {
					break fileLoop
				}
				gd, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range gd.Specs {
					tp, isType := spec.(*ast.TypeSpec)
					if !isType {
						continue
					}

					idents, ok := importTypes[tp.Name.Name]
					if !ok {
						continue
					}
					cnt--
					obj := ast.NewObj(ast.Typ, tp.Name.Name)
					obj.Decl = tp
					for _, ident := range idents {
						ident.Obj = obj
					}
				}
			}
		}
	}

	for model, importedFields := range g.modelImportedFields {
		modelImports := map[string]struct{}{}
		for _, importedField := range importedFields {
			modelImports[importedField.Path] = struct{}{}
			if isFieldRelation(importedField.AstField) {
				importedField.Field.IsSlice = isMany(importedField.AstField.Type)
				importedField.Field.IsElemPointer = isElemPointer(importedField.AstField)
				importedField.Field.IsPointer = isPointer(importedField.AstField)
				model.Relations = append(model.Relations, importedField.Field)
				continue
			}
			g.setModelField(importedField.AstField, importedField.Field, true)
			if model.Primary != importedField.Field {
				model.Fields = append(model.Fields, importedField.Field)
			}
		}

		// Add all imports for given model.
		for imp := range modelImports {
			model.AddImport(imp)
		}
	}
	return nil
}

func (g *ModelGenerator) setModelField(astField *ast.Field, inputField *input.Field, imported bool) {
	isBS, isBSWrapped := isByteSliceWrapper(astField.Type)
	inputField.IsByteSlice = isBS
	inputField.Sortable = isSortable(astField)
	inputField.IsSlice = isMany(astField.Type)
	inputField.IsPointer = isPointer(astField)
	inputField.AlternateTypes = getAlternateTypes(astField.Type)
	setFieldZeroValue(inputField, astField.Type, imported)
	inputField.Selector = getSelector(astField.Type)
	if isBSWrapped {
		inputField.WrappedTypes = []string{"[]byte"}
	} else if !isBS {
		inputField.WrappedTypes = getFieldWrappedTypes(astField)
	}
}

type fieldTag struct {
	key    string
	values []string
}

type importField struct {
	AstField *ast.Field
	Field    *input.Field
	Path     string
	Ident    *ast.Ident
}

func isByteSlice(arr *ast.ArrayType) bool {
	ident, ok := arr.Elt.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "byte"
}

func isSortable(arr *ast.Field) bool {
	if arr.Tag == nil {
		return true
	}
	tags := extractTags(arr.Tag.Value, "neuron", ";", ",")
	for _, tag := range tags {
		switch tag.key {
		case "nosort", "no_sort":
			return false
		}
	}
	return true
}

func isByteSliceWrapper(expr ast.Expr) (isTypeByteSlice bool, isWrapper bool) {
	switch tp := expr.(type) {
	case *ast.ArrayType:
		return isByteSlice(tp), false
	case *ast.Ident:
		if tp.Obj == nil {
			return false, false
		}
		typeSpec, ok := tp.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return false, false
		}
		art, ok := typeSpec.Type.(*ast.ArrayType)
		if !ok {
			return false, false
		}
		ident, ok := art.Elt.(*ast.Ident)
		if !ok {
			return false, false
		}
		return ident.Name == "byte", true
	default:
		return false, false
	}
}

func isExported(field *ast.Field) bool {
	if len(field.Names) == 0 {
		return false
	}
	return field.Names[0].IsExported()
}

func (g *ModelGenerator) isImported(file *ast.File, field *ast.Field) *importField {
	switch tp := field.Type.(type) {
	case *ast.StarExpr:
		sel, isSelector := tp.X.(*ast.SelectorExpr)
		if !isSelector {
			return nil
		}
		return g.createImportField(file, sel)
	case *ast.SelectorExpr:
		return g.createImportField(file, tp)
	default:
		return nil
	}
}

func (g *ModelGenerator) createImportField(file *ast.File, sel *ast.SelectorExpr) *importField {
	pkgIdent, isIdent := sel.X.(*ast.Ident)
	if !isIdent {
		return nil
	}
	i := &importField{
		Ident: sel.Sel,
	}
	for _, imp := range file.Imports {
		p := strings.Trim(imp.Path.Value, "\"")
		base := path.Base(p)
		if base == pkgIdent.Name {
			i.Path = p
			break
		}
	}
	return i
}

func isPrimary(field *ast.Field) bool {
	// Find a neuron primary tag.
	if field.Tag != nil {
		tags := extractTags(field.Tag.Value, "neuron", ";", ",")

		for _, tag := range tags {
			if tag.key == "-" {
				return false
			}
			if strings.EqualFold(tag.key, "type") {
				for _, value := range tag.values {
					switch value {
					case "pk", "primary", "id":
						return true
					}
				}
				break
			}
		}
	}
	// Check if the name suggests it is the "ID" field.
	if strings.EqualFold(field.Names[0].Name, "ID") {
		return true
	}
	return false
}

func isFieldRelation(field *ast.Field) bool {
	return isRelation(field.Type)
}

func isPointer(field *ast.Field) bool {
	_, ok := field.Type.(*ast.StarExpr)
	return ok
}

func isMany(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		if t.Obj == nil {
			return false
		}
		// If this is a wrapper around the slice it would be a type
		dt, ok := t.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return false
		}
		return isMany(dt.Type)
	case *ast.ArrayType:
		return true
	case *ast.StarExpr:
		return isMany(t.X)
	case *ast.SelectorExpr:
		if t.Sel.Obj != nil {
			ts, ok := t.Sel.Obj.Decl.(*ast.TypeSpec)
			if ok {
				return isMany(ts.Type)
			}
		}
		return false
	default:
		return false
	}
}

func isRelation(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		if t.Obj == nil {
			return false
		}
		ts, ok := t.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return false
		}
		return isRelation(ts.Type)
	case *ast.StarExpr:
		return isRelation(t.X)
	case *ast.StructType:
		// Search for the primary key field.
		for _, structField := range t.Fields.List {
			if len(structField.Names) == 0 {
				continue
			}
			name := structField.Names[0]
			if !name.IsExported() {
				continue
			}
			if isPrimary(structField) {
				return true
			}
		}
	case *ast.ArrayType:
		return isRelation(t.Elt)
	case *ast.SelectorExpr:
		return isRelation(t.Sel)
	}
	return false
}

func fieldTypeName(expr ast.Expr) string {
	switch tp := expr.(type) {
	case *ast.Ident:
		return tp.Name
	case *ast.ArrayType:
		switch ln := tp.Len.(type) {
		case *ast.Ellipsis:
			return "[...]" + fieldTypeName(tp.Elt)
		case *ast.BasicLit:
			return "[" + ln.Value + "]" + fieldTypeName(tp.Elt)
		default:
			return "[]" + fieldTypeName(tp.Elt)
		}
	case *ast.StarExpr:
		return "*" + fieldTypeName(tp.X)
	case *ast.StructType:
		name := "struct{"
		for i, field := range tp.Fields.List {
			if len(field.Names) > 0 {
				name += fieldTypeName(field.Names[0]) + " "
			}
			name += fieldTypeName(field.Type)
			if i != len(tp.Fields.List)-1 {
				name += ";"
			}
		}
		name += "}"
		return name
	case *ast.MapType:
		return "map[" + fieldTypeName(tp.Key) + "]" + fieldTypeName(tp.Key)
	case *ast.ChanType:
		switch tp.Dir {
		case ast.RECV:
			return "chan<- " + fieldTypeName(tp.Value)
		case ast.SEND:
			return "<- chan " + fieldTypeName(tp.Value)
		default:
			return "chan " + fieldTypeName(tp.Value)
		}
	case *ast.SelectorExpr:
		return fieldTypeName(tp.X) + "." + tp.Sel.Name
	}
	return ""
}

func extractTags(structTag string, tagName string, tagSeparator, valueSeparator string) []*fieldTag {
	structTag = strings.TrimPrefix(structTag, "`")
	structTag = strings.TrimSuffix(structTag, "`")
	tag, ok := reflect.StructTag(structTag).Lookup(tagName)
	if !ok {
		return nil
	}

	var (
		separators []int
		tags       []*fieldTag
		options    []string
	)

	tagSeparatorRune := []rune(tagSeparator)[0]

	// find all the separators
	for i, r := range tag {
		if i != 0 && r == tagSeparatorRune {
			// check if the  rune before is not an 'escape'
			if tag[i-1] != '\\' {
				separators = append(separators, i)
			}
		}
	}

	// iterate over the option separators
	for i, sep := range separators {
		if i == 0 {
			options = append(options, tag[:sep])
		} else {
			options = append(options, tag[separators[i-1]+1:sep])
		}

		if i == len(separators)-1 {
			options = append(options, tag[sep+1:])
		}
	}
	// if no separators found add the option as whole tag tag
	if options == nil {
		options = append(options, tag)
	}
	// options should be now a legal values defined for the struct tag
	for _, o := range options {
		var equalIndex int
		// find the equalIndex
		for i, r := range o {
			if r == '=' {
				if i != 0 && o[i-1] != '\\' {
					equalIndex = i
					break
				}
			}
		}
		fTag := &fieldTag{}
		if equalIndex != 0 {
			// The left part is the key.
			fTag.key = o[:equalIndex]
			// The right would be the values.
			fTag.values = strings.Split(o[equalIndex+1:], valueSeparator)
		} else {
			// In that case only the key should exists.
			fTag.key = o
		}
		tags = append(tags, fTag)
	}
	return tags
}

func getFieldWrappedTypes(field *ast.Field) []string {
	expr := field.Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	var ident *ast.Ident
	if selector, ok := expr.(*ast.SelectorExpr); ok {
		ident = selector.Sel
	} else if x, ok := field.Type.(*ast.Ident); ok {
		ident = x
	} else {
		return nil
	}
	if ident.Obj == nil {
		return nil
	}
	ts, ok := ident.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		return nil
	}
	return getWrappedTypes(ts.Type)
}

func getWrappedTypes(expr ast.Expr) []string {
	switch x := expr.(type) {
	case *ast.Ident:
		return getWrappedIdent("", x)
	case *ast.SelectorExpr:
		return getWrappedSelector(x)
	// TODO: add case *ast.ArrayType
	default:
		return []string{}
	}
}

func getWrappedSelector(expr *ast.SelectorExpr) []string {
	packageName := fieldTypeName(expr.X)
	return getWrappedIdent(packageName, expr.Sel)
}

func getWrappedIdent(selector string, expr *ast.Ident) []string {
	name := expr.Name
	if selector != "" {
		name = fmt.Sprintf("%s.%s", selector, name)
	}
	if expr.Obj == nil {
		return []string{name}
	}
	ts, ok := expr.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		return []string{name}
	}
	return append([]string{name}, getWrappedTypes(ts.Type)...)
}

func getSelector(expr ast.Expr) string {
	switch tp := expr.(type) {
	case *ast.StarExpr:
		return getSelector(tp.X)
	case *ast.SelectorExpr:
		ident, ok := tp.X.(*ast.Ident)
		if !ok {
			return ""
		}
		return ident.Name
	default:
		return ""
	}
}

func setFieldZeroValue(field *input.Field, expr ast.Expr, imported bool) {
	// Check if given type implements ZeroChecker.
	// TODO: add check if type implements query.ZeroChecker.
	field.Zero = getZeroValue(expr, imported)
	array, ok := expr.(*ast.ArrayType)
	if ok && array.Len == nil {
		field.BeforeZero = "len("
		field.AfterZero = ") == 0"
	} else {
		field.AfterZero = " == " + field.Zero
	}
}

func getZeroValue(expr ast.Expr, imported bool) string {
	switch x := expr.(type) {
	case *ast.Ident:
		if x.Obj == nil {
			switch x.Name {
			case kindInt, kindInt8, kindInt16, kindInt32, kindInt64, kindUint, kindUint8, kindUint16, kindUint32, kindUint64,
				kindByte, kindRune, kindFloat32, kindFloat64:
				return "0"
			case kindString:
				return "\"\""
			case kindBool:
				return "false"
			case kindUintptr:
				return "0"
			case kindPointer:
				return kindNil
			}
		}

		if imported && x.Obj != nil {
			typeSpec, ok := x.Obj.Decl.(*ast.TypeSpec)
			if ok {
				_, isStruct := typeSpec.Type.(*ast.StructType)
				if !isStruct {
					return fmt.Sprintf("%s(%s)", fieldTypeName(expr), getZeroValue(typeSpec.Type, imported))
				}
			}
		}
		return fieldTypeName(expr) + "{}"
	case *ast.StarExpr:
		return kindNil
	case *ast.ArrayType:
		if x.Len == nil {
			// A slice can be nil
			return kindNil
		}
		// The array must be defined to zero values.
		return fieldTypeName(expr) + "{}"
	case *ast.MapType:
		return kindNil
	case *ast.StructType:
		return fieldTypeName(expr) + "{}"
	case *ast.ChanType:
		return kindNil
	case *ast.SelectorExpr:
		selector, ok := x.X.(*ast.Ident)
		if !ok {
			return fieldTypeName(expr) + "{}"
		}
		return selector.Name + "." + getZeroValue(x.Sel, imported)
	default:
		return kindNil
	}
}

func getAlternateTypes(expr ast.Expr) []string {
	switch exprType := expr.(type) {
	case *ast.Ident:
		return getIdentAlternateTypes(exprType)
	case *ast.ArrayType:
		return getArrayAlternateTypes(exprType)
	case *ast.SelectorExpr:
		return getIdentAlternateTypes(exprType.Sel)
	case *ast.StarExpr:
		return getAlternateTypes(exprType.X)
	}
	return []string{}
}

func getArrayAlternateTypes(expr *ast.ArrayType) []string {
	switch expr.Len.(type) {
	case *ast.BasicLit: // Array
		return nil
	case *ast.Ellipsis: // [...] Array
		return nil
	default: // Slice
		// check if this is a byte slice
		ident, ok := expr.Elt.(*ast.Ident)
		if !ok {
			return nil
		}
		if ident.Obj != nil {
			return nil
		}
		if ident.Name == kindByte {
			return []string{"string"}
		}
		return nil
	}
}

func getIdentAlternateTypes(expr *ast.Ident) []string {
	if expr.Obj == nil {
		return getBasicAlternateTypes(expr)
	}
	ot, ok := expr.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		return []string{}
	}
	return getAlternateTypes(ot.Type)
}

func getBasicAlternateTypes(expr *ast.Ident) (alternateTypes []string) {
	switch expr.Name {
	case kindInt, kindInt8, kindInt16, kindInt32, kindInt64, kindUint, kindUint8, kindUint16, kindUint32, kindUint64,
		kindByte, kindRune, kindFloat32, kindFloat64:
		for _, kind := range []string{kindInt, kindInt8, kindInt16, kindInt32, kindInt64, kindUint, kindUint8, kindUint16, kindUint32, kindUint64,
			kindFloat32, kindFloat64} {
			if kind != expr.Name {
				alternateTypes = append(alternateTypes, kind)
			}
		}
	case kindString:
		alternateTypes = []string{"[]byte"}
	}
	return alternateTypes
}
