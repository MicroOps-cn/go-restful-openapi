package restfulspec

import (
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/gogo/protobuf/proto"
)

func initPropExtensions(ext *spec.Extensions) {
	if *ext == nil {
		*ext = make(spec.Extensions, 0)
	}
}

func setDescription(prop *spec.Schema, field reflect.StructField) {
	if tag := field.Tag.Get("description"); tag != "" {
		prop.Description = tag
	}
}

func setDefaultValue(prop *spec.Schema, field reflect.StructField) {
	if tag := field.Tag.Get("default"); tag != "" {
		prop.Default = stringAutoType(tag)
	}
}

func setIsNullableValue(prop *spec.Schema, field reflect.StructField) {
	if tag := field.Tag.Get("x-nullable"); tag != "" {
		initPropExtensions(&prop.Extensions)

		value, err := strconv.ParseBool(tag)

		prop.Extensions["x-nullable"] = value && err == nil
	}
}

func setGoNameValue(prop *spec.Schema, field reflect.StructField) {
	const tagName = "x-go-name"
	if tag := field.Tag.Get(tagName); tag != "" {
		initPropExtensions(&prop.Extensions)
		prop.Extensions[tagName] = tag
	}
}

type EnumItem struct {
	name  string
	value int32
}

type EnumItems []EnumItem

func (e EnumItems) Len() int {
	return len(e)
}

func (e EnumItems) Less(i, j int) bool {
	return e[i].value < e[j].value
}

func (e EnumItems) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func setEnumValues(b definitionBuilder, prop *spec.Schema, field reflect.StructField) {
	// We use | to separate the enum values.  This value is chosen
	// since it's unlikely to be useful in actual enumeration values.
	if tag := field.Tag.Get("enum"); tag != "" {
		var enums []interface{}
		for _, s := range strings.Split(tag, "|") {
			enums = append(enums, s)
		}
		prop.Enum = enums
		return
	} else if protoTag := field.Tag.Get("protobuf"); protoTag != "" {
		var typeName string
		var hasRep bool

		for _, s := range strings.Split(protoTag, ",") {
			if s == "rep" {
				hasRep = true
			}
			if strings.HasPrefix(s, "enum=") {
				typeName = s[5:]
				break
			}
		}
		if len(typeName) != 0 {
			enumMap := proto.EnumValueMap(typeName)
			var enumItems EnumItems
			for name, val := range enumMap {
				enumItems = append(enumItems, EnumItem{value: val, name: name})
			}
			sort.Sort(enumItems)
			var enums = make([]interface{}, 0)
			for _, item := range enumItems {
				enums = append(enums, item.name)
				enums = append(enums, item.value)
			}
			if _, ok := b.Definitions[typeName]; !ok {
				schema := spec.Schema{}
				schema.Enum = enums
				b.Definitions[typeName] = schema
			}
			if hasRep {
				prop.Type = spec.StringOrArray{"array"}
				prop.Items = &spec.SchemaOrArray{
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Ref: spec.MustCreateRef("#/definitions/" + typeName),
						},
					},
				}
			} else {
				prop.Ref = spec.MustCreateRef("#/definitions/" + typeName)
			}
		}
	}
}

func setFormat(prop *spec.Schema, field reflect.StructField) {
	if tag := field.Tag.Get("format"); tag != "" {
		prop.Format = tag
	}

}

func setMaximum(prop *spec.Schema, field reflect.StructField) {
	if tag := field.Tag.Get("maximum"); tag != "" {
		value, err := strconv.ParseFloat(tag, 64)
		if err == nil {
			prop.Maximum = &value
		}
	}
}

func setMinimum(prop *spec.Schema, field reflect.StructField) {
	if tag := field.Tag.Get("minimum"); tag != "" {
		value, err := strconv.ParseFloat(tag, 64)
		if err == nil {
			prop.Minimum = &value
		}
	}
}

func setPattern(prop *spec.Schema, field reflect.StructField) {
	if tag := field.Tag.Get("pattern"); tag != "" {
		prop.Pattern = tag
	}
}
func setType(prop *spec.Schema, field reflect.StructField) {
	if tag := field.Tag.Get("type"); tag != "" {
		// Check if the first two characters of the type tag are
		// intended to emulate slice/array behaviour.
		//
		// If type is intended to be a slice/array then add the
		// overridden type to the array item instead of the main property
		if len(tag) > 2 && tag[0:2] == "[]" {
			pType := "array"
			prop.Type = []string{pType}
			prop.Items = &spec.SchemaOrArray{
				Schema: &spec.Schema{},
			}
			iType := tag[2:]
			prop.Items.Schema.Type = []string{iType}
			return
		}

		prop.Type = []string{tag}
	}
}

func setUniqueItems(prop *spec.Schema, field reflect.StructField) {
	tag := field.Tag.Get("unique")
	switch tag {
	case "true":
		prop.UniqueItems = true
	case "false":
		prop.UniqueItems = false
	}
}

func setReadOnly(prop *spec.Schema, field reflect.StructField) {
	tag := field.Tag.Get("readOnly")
	switch tag {
	case "true":
		prop.ReadOnly = true
	case "false":
		prop.ReadOnly = false
	}
}

func setPropertyMetadata(b definitionBuilder, prop *spec.Schema, field reflect.StructField) {
	setDescription(prop, field)
	setDefaultValue(prop, field)
	setEnumValues(b, prop, field)
	setFormat(prop, field)
	setMinimum(prop, field)
	setMaximum(prop, field)
	setPattern(prop, field)
	setUniqueItems(prop, field)
	setType(prop, field)
	setReadOnly(prop, field)
	setIsNullableValue(prop, field)
	setGoNameValue(prop, field)
}
