package reflectgen

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"go/format"
	"hash"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/neuronlabs/errors"

	"github.com/neuronlabs/neuron/class"
	"github.com/neuronlabs/neuron/mapping"
	"github.com/neuronlabs/neuron/namer"
)

// Config is the config input for the generated files.
type Config struct {
	PackageName string
	Path        string
}

// Generator is the neuron model generator.
type Generator struct {
	Config    *Config
	Models    []*mapping.ModelStruct
	Templates *template.Template
}

// New creates new generator for provided config 'cfg' and 'models'.
func New(cfg *Config, models ...*mapping.ModelStruct) (*Generator, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	if cfg.Path == "" {
		cfg.Path = "./models"
	}
	if cfg.PackageName == "" {
		cfg.PackageName = filepath.Dir(cfg.Path)
	}

	if len(models) == 0 {
		return nil, errors.New(class.GeneratorInvalidInput, "no models provided")
	}

	gen := &Generator{
		Config: cfg,
		Models: models,
	}

	// Read the templates and compile the templates.
	gen.Templates = template.New("")
	for _, tmpl := range bintemplates.AssetNames() {
		data, err := bintemplates.Asset(tmpl)
		if err != nil {
			return nil, err
		}
		_, err = gen.Templates.New(tmpl).Parse(string(data))
		if err != nil {
			return nil, err
		}
	}
	return gen, nil
}

// Generate creates or updates files for all models as well as an initialization.
func (g *Generator) Generate() (err error) {
	buf := &bytes.Buffer{}
	h := md5.New()
	for _, model := range g.Models {
		if err = g.generateModel(model, buf, h); err != nil {
			return err
		}
	}
	return nil
}

// nolint:errcheck
func (g *Generator) generateModel(model *mapping.ModelStruct, buf *bytes.Buffer, h hash.Hash) error {
	input, err := g.createModelInput(model)
	if err != nil {
		return err
	}
	// Write at first to buffer.
	if err = g.Templates.ExecuteTemplate(buf, "model", input); err != nil {
		return err
	}

	// gofmt generated file.
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	buf.Reset()

	if _, err = h.Write(formatted); err != nil {
		return err
	}
	genCheckSum := h.Sum(nil)
	h.Reset()

	// Open the file and check if the hash matches.
	fileName := filepath.Join(g.Config.Path, namer.NamingKebab(model.Collection())+".go")
	writeToFile := true
	f, err := os.Open(fileName)
	if err != nil {
		if !strings.Contains(err.Error(), "no such file or directory") {
			return err
		}
		// Create a new file with given name.
		f, err = os.Create(fileName)
		if err != nil {
			return err
		}
	} else {
		// If the file exists check if the md5 check sum matches.
		if _, err = io.Copy(h, f); err != nil {
			f.Close()
			return err
		}
		existingCheckSum := h.Sum(nil)
		writeToFile = bytes.Equal(existingCheckSum, genCheckSum)
	}
	defer f.Close()

	if !writeToFile {
		return nil
	}
	_, err = f.Write(formatted)
	if err != nil {
		return err
	}
	return nil
}

func (g *Generator) createModelInput(mStruct *mapping.ModelStruct) (*ModelTemplateInput, error) {
	input := &ModelTemplateInput{
		PackageName: g.Config.PackageName,
		Collection: Collection{
			Name:         mStruct.Collection(),
			Receiver:     strings.ToLower(mStruct.Type().Name()[:1]),
			VariableName: namer.NamingCamel(mStruct.Collection()),
			QueryBuilder: mStruct.Collection() + "Builder",
		},
		Imports: []string{"github.com/neuronlabs/errors", "github.com/neuronlabs/neuron/class", "github.com/neuronlabs/neuron/query"},
	}
	model := Model{
		Name:       mStruct.Type().Name(),
		Receiver:   strings.ToLower(mStruct.Type().Name()[:1]),
		Collection: mStruct.Collection(),
	}

	var (
		scanner     sql.Scanner
		zeroChecker mapping.ZeroChecker
	)
	scannerType := reflect.TypeOf(scanner)
	zeroCheckerType := reflect.TypeOf(zeroChecker)
	for _, field := range mStruct.Fields() {
		fieldType := field.ReflectField().Type
		pkgPath := fieldType.PkgPath()
		// Add required imports for given model.
		if pkgPath != "" {
			input.addImport(pkgPath)
		}
		genField := &Field{
			Name:           field.Name(),
			Sortable:       field.CanBeSorted(),
			Type:           fieldType.String(),
			AlternateTypes: nil,
			Scanner:        fieldType.Implements(scannerType),
			ZeroChecker:    fieldType.Implements(zeroCheckerType),
			NeuronField:    field,
		}

		switch fieldType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float64, reflect.Float32:
			intKinds := []reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float32, reflect.Float64}
			for _, intKind := range intKinds {
				if intKind != fieldType.Kind() {
					genField.AlternateTypes = append(genField.AlternateTypes, intKind.String())
				}
			}
			genField.AlternateTypes = append(genField.AlternateTypes, "byte", "rune")
		case reflect.String:
			genField.AlternateTypes = []string{"[]byte"}
		}
		if fieldType.String() == "[]byte" {
			genField.AlternateTypes = []string{"string"}
		}

		if !genField.ZeroChecker {
			switch fieldType.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float32, reflect.Float64:
				genField.Zero = "0"
			case reflect.String:
				genField.Zero = "\"\""
			case reflect.Bool:
				genField.Zero = "false"
			case reflect.Ptr, reflect.Map:
				genField.Zero = "nil"
			case reflect.Slice:
				genField.Zero = "nil"
				genField.BeforeZero = "len("
				genField.AfterZero = ") == 0"
			case reflect.Struct:
				genField.Zero = fieldType.String() + "{}"
			}
			if genField.AfterZero == "" {
				genField.AfterZero = " == " + genField.Zero
			}
		}

		if !genField.ZeroChecker && genField.Zero == "" {
			return nil, errors.Newf(class.GeneratorFieldNoZeroer, "provided genField: '%s' have no zero value", genField.Name)
		}

		switch field.Kind() {
		case mapping.KindPrimary:
			model.Primary = genField
		case mapping.KindForeignKey:
			model.Attributes = append(model.Attributes, genField)
		case mapping.KindAttribute:
			model.ForeignKeys = append(model.ForeignKeys, genField)
		default:
			return nil, errors.Newf(class.Internal, "invalid genField: '%s' kind: '%s'", field.Name(), field.Kind())
		}
		model.Fields = append(model.Fields, genField)
	}

	model.Fielder = len(model.Attributes) != 0 || len(model.ForeignKeys) != 0

	for _, relation := range mStruct.RelationFields() {
		relationField := &Field{
			Name:        relation.Name(),
			Type:        relation.ReflectField().Type.Name(),
			NeuronField: relation,
		}
		switch relation.Kind() {
		case mapping.KindRelationshipMultiple:
			model.MultiRelationer = true
		case mapping.KindRelationshipSingle:
			model.SingleRelationer = true
		}
		model.Relations = append(model.Relations, relationField)
	}

	for _, privateField := range mStruct.PrivateFields() {
		private := &Field{
			Name:        privateField.Name(),
			Type:        privateField.ReflectField().Type.Name(),
			NeuronField: privateField,
			Tags:        string(privateField.ReflectField().Tag),
		}
		model.PrivateFields = append(model.PrivateFields, private)
	}

	// Recompute field indexes.
	var indexCount int
	model.Primary.Index = indexCount
	indexCount++

	for _, attr := range model.Attributes {
		attr.Index = indexCount
		indexCount++
	}
	for _, relation := range model.Relations {
		relation.Index = indexCount
		indexCount++
	}

	for _, foreignKey := range model.ForeignKeys {
		foreignKey.Index = indexCount
		indexCount++
	}

	for _, private := range model.PrivateFields {
		private.Index = indexCount
		indexCount++
	}
	return input, nil
}

//
// Examples:
//
// var users *usersCollection
// func Users() *usersCollection {
//
// }
//
// Users.Where("ID > 10").Find()
// Users.Model(u).SetCars()
//
//
// type User struct {
//		ID		int
//		Name	string
//		House 	*House
//		HouseID *int
//		Pets 	[]*Pet
// }
//
// type Pet struct {
//		ID			int
//		Owner		*User
//		OwnerID		int
// }
//
// type House struct {
//		ID 		int
//		Inhabitants []*User
//		Address	string
// }
//	NOTE: the 'Users' should take ...*User input.
// models.Users(u).
// 	Select("ID", "Name", "HouseID").
// 	Insert(ctx, tx)
//
// users, err := models.Users().
// 	Select("ID", "Name").
//	Where("ID > ", 10).
//	Limit(10).
//	Offset(10).
//	Include("House").
// 	Find(ctx, db)
//
//
//
// models.Users(u1, u2, u3).SetPets(pets)
// models.Users(u1, u2, u3).SetHouse(
// models.Users(u1, u2, u3).SetRelationships(relField, []query.Model{})

//	u.SetHouse()
//	u.RemoveHouse(ctx, db)
//
//	u.AddPets(pets...)
//
// ToMany relations
// Query(model, modelInst1, modelInst2).
//	AddRelationships(relField, relModels...)	- add 'model' relationships
//	SetRelationships(relField, relModels...)	- clear relationships and set again.
//	RemoveRelationships(relField, relModel...) 	- remove provided relationships
//
//	ToOne
//	Query(model, modelInst).
//	SetRelationship(relField, relModel) 		- set 'relModel' to the 'relField'
//	RemoveRelationship(relField)				- remove modelInst 'relField'

//
// Separate query processes.
//
//	Find()	([]query.Model, error)
//	NOTE: this function should return error if the model is not found.
//	Get()	(query.Model, error)
//	NOTE: - this should take input models and refresh their values for selected fields - shouldn't allow filters.
//	Refresh() error
//	BREAKING CHANGE: Change 'Insert' into 'Insert'
//	Insert() error
//	BREAKING CHANGE: Change 'Update' into 'Update'
//	Update() error
//	NOTE: Insert and on conflict - update the model.
//	Upsert() error
//	BREAKING CHANGE: Delete should now allow only a single model instance to delete without filters.
//	Delete() error
//	NOTE: Query defined functions doesn't require model values - they operate on filters and limits.
//	Count() (int64, error)
//	Exists() (bool, error)
//	NOTE: QueryDelete and QueryUpdate are separate processes just to make it transparent how it works. This should
//		allow filters (and fields for Update) to affect multiple instances - the number of affected should be returned
//	QueryDelete() (int64, error)
//	QueryUpdate() (int64, error)
//
//	NOTE: BatchInsert and BatchUpdate may use defined fieldset - otherwise they take zero values for each instance.
//		A user defines transaction - no auto - transaction at all.
//	BatchInsert() error
//	BatchUpdate() error
//	BatchDelete() error
//
//
// Also relationship fields changers should be separate processes to make it easier.
//	NOTE: For both addRelations and SetRelations if there are many input model instances
//		The relation must be of many2many type - otherwise it is not possible for multiple models.
//		All these methods must operate on MultipleRelationships.
//	addRelations(relField *mapping.StructField, models ...query.Model) error
// 	SetRelations(relField *mapping.StructField, models ...query.Model) error
//	RemoveRelations(relField *mapping.StructField)
//
//	NOTE: these methods are used only for the single relationships like has-one or belongs to.
//		For multiple models input the relationship must be of 'belongs-to' type.
//	SetRelation(relField *mapping.StructField, model query.Model) error
//	RemoveRelation(relField *mapping.StructField)
//
//

//
// Repository methods
//
//	Insert
//	Update
//	Upsert
//	Delete
//	Count
//	Exists
//	Find

//
// DB interface
//
// NOTE: Batch functions should be used if length of the models is bigger than one - otherwise use singular form.

//
//	NeuronCollectionName generated methods
//
// NOTE: Collections should not use 'BatchNamings' as they.
// BREAKING CHANGE: 'Insert' changed the name into 'Insert'.
// Insert(ctx, db) error - get input model
// NOTE: the find function should return ModelSlice we can use some methods for the modelSlice.
// Find(ctx, db) (ModelSlice, error)
// Get(ctx, db)	(*Model, error)
// NOTE: Added new method
// Refresh(ctx, db) error
// BREAKING CHANGE: 'Update' changed name to 'Update'
// Update(ctx, db) error
// NOTE: Added new method.
// Upsert(ctx, db) error
// NOTE: Added new method.
// Exists(ctx, db) (bool, error)
// Count(ctx, db) (int64, error)
//
// Query Builder methods
//
// BREAKING CHANGE: 'Where' changed the name into 'Where'.
// Where
// BREAKING CHANGE: fields should only allow primaries, attributes and foreign keys. Relationships no longer should be
//		allowed by the Select method.
// Select
// BREAKING CHANGE: PageSize and PageNumber should be deleted from query.Pagination.
// Limit, Offset
// Include
// OrderBy
//
// NOTE: If the model is MultiRelationer - implement all functions of Multi relationships with theirs names and valid field model.
// NOTE: If the model is SingleRelationer - implement all functions of single relationships with theirs names and valid field model.
