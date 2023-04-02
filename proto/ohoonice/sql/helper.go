package sql

import (
	"github.com/envoyproxy/protoc-gen-validate/validate"
	"github.com/srikrsna/protoc-gen-gotag/tagger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func GetTable(field protoreflect.MessageDescriptor) *Table {
	if field == nil {
		return nil
	}
	v, ok := proto.GetExtension(field.Options(), E_Table).(*Table)
	if v != nil && ok {
		return v
	}
	return nil
}

func GetValidateFieldRule(field protoreflect.FieldDescriptor) *validate.FieldRules {
	if field == nil {
		return nil
	}
	v, ok := proto.GetExtension(field.Options(), validate.E_Rules).(*validate.FieldRules)
	if v != nil && ok {
		return v
	}
	return nil
}

func GetTag(field protoreflect.FieldDescriptor) string {
	if field == nil {
		return ""
	}
	v, ok := proto.GetExtension(field.Options(), tagger.E_Tags).(string)
	if ok {
		return v
	}
	return ""
}

func (t *Table) ExtractKeyFields() []string {
	var m = map[string]bool{}
	for _, f := range t.GetPrimaryKey().GetF() {
		m[f] = true
	}

	for i := range t.GetUniqueKeys() {
		for _, f := range t.GetUniqueKeys()[i].GetF() {
			m[f] = true
		}
	}

	for i := range t.GetKeys() {
		for _, f := range t.GetKeys()[i].GetF() {
			m[f] = true
		}
	}

	ret := make([]string, 0, len(m))
	for f := range m {
		ret = append(ret, f)
	}

	return ret
}
