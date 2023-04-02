package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sql2 "github.com/ohoonice/protoc-gen-sql/proto/ohoonice/sql"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GenerateFile generates a _http.pb.go file containing doom errors definitions.
func GenerateFile(gen *protogen.Plugin, file *protogen.File, outdir, database string) {

	for _, m := range file.Messages {
		rule := sql2.GetTable(m.Desc)
		if rule == nil {
			continue
		}

		table := &tableDesc{
			Database: database,
			Table:    rule.GetTable(),
		}

		fs := rule.ExtractKeyFields()
		for _, f := range fs {
			notFound := true
			for _, mf := range m.Fields {
				if f == string(mf.Desc.Name()) {
					notFound = false
					break
				}
			}
			if notFound {
				gen.Error(fmt.Errorf("not found key field: %s", f))
				return
			}
		}

		pfm := make(map[string]bool)
		for _, f := range rule.GetPrimaryKey().GetF() {
			pfm[f] = true
		}
		nameMap := make(map[string]string)

		for _, f := range m.Fields {
			fs, name, err := GetFieldSQL(pfm, f.Desc)
			if err != nil {
				gen.Error(err)
				return
			}
			if name != "" {
				nameMap[string(f.Desc.Name())] = name
			}

			table.Fields = append(table.Fields, fs)
		}

		renameKeyF(nameMap, rule.GetPrimaryKey())
		renameKeysF(nameMap, rule.GetUniqueKeys())
		renameKeysF(nameMap, rule.GetKeys())

		table.PrimaryKey = convertPrimaryKey(rule.GetPrimaryKey())
		table.UniqueKeys = batchConvertUniqueKey(rule.GetUniqueKeys())
		table.Keys = batchConvertKey(rule.GetKeys())

		filename := table.Table + ".sql"

		f, err := os.OpenFile(filepath.Join(outdir, filename), os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			gen.Error(err)
		}

		_, err = f.WriteString(table.execute())
		if err != nil {
			gen.Error(err)
		}
	}
}

func renameKeysF(nameMap map[string]string, keys []*sql2.Key) {
	for _, k := range keys {
		renameKeyF(nameMap, k)
	}
}

func renameKeyF(nameMap map[string]string, key *sql2.Key) {
	for i, f := range key.F {
		if name, ok := nameMap[f]; ok {
			key.F[i] = name
		}
	}
}

func convertPrimaryKey(key *sql2.Key) string {
	// uk_platform_vid(`platform`,`vid`)
	//name := strings.Join(key.F, "_")
	quotes := make([]string, 0, len(key.F))
	for _, f := range key.F {
		quotes = append(quotes, "`"+f+"`")
	}
	domain := strings.Join(quotes, ",")
	return fmt.Sprintf("(%s)", domain)
}

func batchConvertUniqueKey(keys []*sql2.Key) []string {
	ret := make([]string, 0, len(keys))
	for _, key := range keys {
		ret = append(ret, convertUniqueKey(key))
	}
	return ret
}

func convertUniqueKey(key *sql2.Key) string {
	// uk_platform_vid(`platform`,`vid`)
	name := strings.Join(key.F, "_")
	quotes := make([]string, 0, len(key.F))
	for _, f := range key.F {
		quotes = append(quotes, "`"+f+"`")
	}
	domain := strings.Join(quotes, ",")
	return fmt.Sprintf("uk_%s(%s)", name, domain)
}

func batchConvertKey(keys []*sql2.Key) []string {
	ret := make([]string, 0, len(keys))
	for _, key := range keys {
		ret = append(ret, convertKey(key))
	}
	return ret
}

func convertKey(key *sql2.Key) string {
	// uk_platform_vid(`platform`,`vid`)
	name := strings.Join(key.F, "_")
	quotes := make([]string, 0, len(key.F))
	for _, f := range key.F {
		quotes = append(quotes, "`"+f+"`")
	}
	domain := strings.Join(quotes, ",")
	return fmt.Sprintf("idx_%s(%s)", name, domain)
}

func GetFieldSQL(pfm map[string]bool, field protoreflect.FieldDescriptor) (string, string, error) {
	isPrimaryKey := false
	if _, ok := pfm[string(field.Name())]; ok && len(pfm) == 1 {
		isPrimaryKey = true
	}
	name := string(field.Name())

	tag := sql2.GetTag(field) //"gorm:\"column:id\""

	if len(tag) != 0 {
		tags := strings.Split(tag, " ")
		for _, t := range tags {
			fs := strings.Split(tag, `"`)
			if len(fs) != 3 {
				return "", "", fmt.Errorf("tag is invalid: tagField: [%s], tag: [%s], fieldName: [%s]", t, tag, field.Name())
			}

			if fs[0] == "gorm:" {
				ns := strings.Split(fs[1], ":")
				if len(ns) != 2 {
					return "", "", fmt.Errorf("gorm tag is invalid: %s", tag)
				}
				if ns[0] == "column" {
					name = ns[1]
				}
			}
		}
	}

	switch field.Kind() {
	case protoreflect.BoolKind:
		return fmt.Sprintf("%s BOOL NOT NULL DEFAULT false,", name), name, nil
	case protoreflect.Int32Kind:
		if isPrimaryKey {
			return fmt.Sprintf("%s INT NOT NULL AUTO_INCREMENT,", name), name, nil
		}
		return fmt.Sprintf("%s INT NOT NULL DEFAULT 0,", name), name, nil
	case protoreflect.Int64Kind:
		if isPrimaryKey {
			return fmt.Sprintf("%s BIGINT NOT NULL AUTO_INCREMENT,", name), name, nil
		}
		return fmt.Sprintf("%s BIGINT NOT NULL DEFAULT 0,", name), name, nil
	case protoreflect.Uint32Kind:
		if isPrimaryKey {
			return fmt.Sprintf("%s INT UNSIGNED NOT NULL AUTO_INCREMENT,", name), name, nil
		}
		return fmt.Sprintf("%s INT UNSIGNED NOT NULL DEFAULT 0,", name), name, nil
	case protoreflect.Uint64Kind:
		if isPrimaryKey {
			return fmt.Sprintf("%s BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,", name), name, nil
		}
		return fmt.Sprintf("%s BIGINT UNSIGNED NOT NULL DEFAULT 0,", name), name, nil
	case protoreflect.EnumKind:
		values := field.Enum().Values()
		enums := make([]string, 0, values.Len())
		for i := 0; i < values.Len(); i++ {
			desc := values.Get(i)
			enums = append(enums, fmt.Sprintf("%d=%s", desc.Number(), desc.Name()))
		}
		comment := strings.Join(enums, "; ")
		return fmt.Sprintf("%s INT NOT NULL DEFAULT 0 COMMENT '%s',", name, comment), name, nil
	case protoreflect.FloatKind:
		return fmt.Sprintf("%s FLOAT NOT NULL DEFAULT 0,", name), name, nil
	case protoreflect.DoubleKind:
		return fmt.Sprintf("%s DOUBLE NOT NULL DEFAULT 0,", name), name, nil
	case protoreflect.StringKind:
		f := sql2.GetValidateFieldRule(field)
		if f == nil {
			return "", "", fmt.Errorf("no max_len validation for field: %s", name)
		}
		fieldMaxLen := f.GetString_().GetMaxLen()
		if fieldMaxLen == 0 {
			return "", "", fmt.Errorf("no max_len validation for field: %s", name)
		} else if fieldMaxLen <= 1024 {
			return fmt.Sprintf("%s VARCHAR(%d) NOT NULL DEFAULT '',", name, fieldMaxLen), name, nil
		} else {
			return fmt.Sprintf("%s TEXT,", name), name, nil
		}
	case protoreflect.BytesKind:
		return fmt.Sprintf("%s TEXT,", name), name, nil
	case protoreflect.MessageKind:
		return fmt.Sprintf("%s JSON,", name), name, nil
	}

	return "", "", fmt.Errorf("not support field kind: %s", field.Name())
}
