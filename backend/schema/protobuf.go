package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var typesWithRuntime = map[string]bool{
	"Module": true,
	"Verb":   true,
}

// ProtobufSchema returns a string containing the equivalent protobuf schema
// for the FTL schema.
func ProtobufSchema() string {
	messages := map[string]string{}
	generateMessage(reflect.TypeOf(Schema{}), messages)
	keys := maps.Keys(messages)
	slices.Sort(keys)
	w := &strings.Builder{}
	fmt.Fprintf(w, `syntax = "proto3";

// This file is generated by github.com/TBD54566975/ftl/backend/schema/protobuf.go, DO NOT MODIFY
package xyz.block.ftl.v1.schema;

import "xyz/block/ftl/v1/schema/runtime.proto";

option go_package = "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema;schemapb";
option java_multiple_files = true;
`)
	for i, key := range keys {
		w.WriteString(messages[key])
		if i != len(keys)-1 {
			w.WriteString("\n")
		}
	}
	return w.String()
}

func reflectUnion[T any](union ...T) []reflect.Type {
	out := []reflect.Type{}
	for _, t := range union {
		out = append(out, reflect.TypeOf(t))
	}
	return out
}

func generateMessage(et reflect.Type, messages map[string]string) {
	et = indirect(et)
	if et.Kind() == reflect.Interface {
		generateUnion(et, messages)
	} else {
		generateStruct(et, messages)
	}
}

func generateStruct(t reflect.Type, messages map[string]string) {
	t = indirect(t)
	if _, ok := messages[t.Name()]; ok {
		return
	}
	messages[t.Name()] = ""
	w := &strings.Builder{}
	fmt.Fprintf(w, "\nmessage %s {", t.Name())
	if typesWithRuntime[t.Name()] {
		fmt.Fprintf(w, "\n  optional %sRuntime runtime = 31634;\n", t.Name())
	}
	fields := reflect.VisibleFields(t)
	// Sort by protobuf tag
	slices.SortFunc(fields, func(a, b reflect.StructField) int {
		aid := strings.Split(a.Tag.Get("protobuf"), ",")[0]
		bid := strings.Split(b.Tag.Get("protobuf"), ",")[0]
		if aid == "-" || bid == "-" {
			return strings.Compare(aid, bid)
		}
		an, _ := strconv.Atoi(aid)
		bn, _ := strconv.Atoi(bid)
		return an - bn
	})
	// Filter out fields with protobuf tag "-"
	filtered := []reflect.StructField{}
	for _, ft := range fields {
		tag := strings.Split(ft.Tag.Get("protobuf"), ",")
		if tag[0] == "-" {
			continue
		}
		filtered = append(filtered, ft)
	}
	if len(filtered) == 0 {
		fmt.Fprint(w, "}")
	} else {
		for _, ft := range filtered {
			ftt := indirect(ft.Type)
			tag := strings.Split(ft.Tag.Get("protobuf"), ",")
			_, err := strconv.Atoi(tag[0])
			if err != nil {
				panic(fmt.Sprintf("%s.%s: invalid protobuf tag %q", t.Name(), ft.Name, tag[0]))
			}
			opt := ""
			if len(tag) > 1 && tag[1] == "optional" {
				opt = "optional "
			}
			fmt.Fprintf(w, "\n  %s%s %s = %s;", opt, generateProtoType(ft.Type), strcase.ToLowerCamel(ft.Name), tag[0])
			et := elemType(ftt)
			if et.Kind() == reflect.Struct || et.Kind() == reflect.Interface {
				generateMessage(et, messages)
			}
		}
		fmt.Fprint(w, "\n}")
	}
	messages[t.Name()] = w.String()
}

func generateUnion(t reflect.Type, messages map[string]string) {
	t = indirect(t)
	if _, ok := messages[t.Name()]; ok {
		return
	}
	messages[t.Name()] = ""
	if _, ok := unions[t]; !ok {
		panic("no union defined for " + t.Name())
	}
	w := &strings.Builder{}
	fmt.Fprintf(w, "\nmessage %s {\n", t.Name())
	fmt.Fprintln(w, "  oneof value {")
	for i, ut := range unions[t] {
		ut = indirect(ut)
		fmt.Fprintf(w, "    %s %s = %d;\n", ut.Name(), strcase.ToLowerCamel(strings.TrimPrefix(ut.Name(), t.Name())), i+1)
		generateMessage(ut, messages)
	}
	fmt.Fprintln(w, "  }")
	fmt.Fprint(w, "}")
	messages[t.Name()] = w.String()
}

func generateProtoType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Ptr:
		return generateProtoType(t.Elem())
	case reflect.Interface:
		return t.Name()
	case reflect.Struct:
		return t.Name()
	case reflect.String:
		return "string"
	case reflect.Int:
		return "int64"
	case reflect.Bool:
		return "bool"
	case reflect.Float64:
		return "double"
	case reflect.Slice:
		return fmt.Sprintf("repeated %s", generateProtoType(t.Elem()))
	case reflect.Map:
		return fmt.Sprintf("map<string, %s>", generateProtoType(t.Elem()))
	default:
		panic(fmt.Sprintf("unsupported type: %s", t))
	}
}

func elemType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Map {
		return elemType(t.Elem())
	}
	return t
}

func indirect(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}
