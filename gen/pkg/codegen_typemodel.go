package glot

import (
	"strconv"
	"strings"
)

type GenEnumeration struct {
	GenBase
	Type       GenType
	Enumerants []GenEnumerant
}
type GenEnumerant struct {
	GenBase
	Value string
}

type GenStructure struct {
	GenBase
	Extends    []GenType
	Mixins     []GenType
	Properties []GenProperty
}
type GenProperty struct {
	GenBase
	Type     GenType
	Optional bool
}

type GenBase struct {
	Name       string
	NameUp     string
	DocLines   []string
	Since      string
	Deprecated string
}

func (it *GenBase) base() *GenBase { return it }

func (it *GenBase) DocComments(root *GenDot) string {
	if len(it.DocLines) == 0 {
		return ""
	}
	doc_lines := Copy(it.DocLines)
	for i, doc_line := range doc_lines {
		for idx1 := strings.Index(doc_line, "{@link "); idx1 >= 0; idx1 = strings.Index(doc_line, "{@link ") {
			if idx2 := strings.Index(doc_line[idx1:], "}"); idx2 < 0 {
				break
			} else {
				link_strs := []string{doc_line[idx1+len("{@link ") : idx2+idx1]}
				if idx_space := strings.IndexByte(link_strs[0], ' '); idx_space > 0 {
					link_strs = append(link_strs, link_strs[0][idx_space+1:])
					link_strs[0] = link_strs[0][:idx_space]
				}
				rendered := root.gen.tmplExec(nil, root.gen.tmpl("doc_comments_link", "`{{index . 0}}`"), link_strs)
				doc_lines[i] = doc_line[:idx1] + If(rendered == "", link_strs[0], rendered) + doc_line[idx2+idx1+1:]
				doc_line = doc_lines[i]
			}
		}
	}
	if it.Since != "" && !Exists(doc_lines, func(s string) bool { return strings.Contains(s, "@since") }) {
		doc_lines = append(doc_lines, "@since "+it.Since)
	}
	if it.Deprecated != "" && !Exists(doc_lines, func(s string) bool { return strings.Contains(s, "@deprecated") }) {
		doc_lines = append(doc_lines, "@deprecated "+it.Deprecated)
	}
	return root.gen.tmplExec(nil, root.gen.tmpl("doc_comments", ""), doc_lines)
}

type GenType interface {
	key() string
	kind() string
}

func genTypeString(it GenType) string {
	return "<" + it.kind() + ">" + it.kind()
}

type GenTypeBaseType string

func (it GenTypeBaseType) String() string { return Up0(string(it)) }
func (it GenTypeBaseType) kind() string   { return "Base" }
func (it GenTypeBaseType) key() string    { return "b#" + string(it) }

type GenTypeEnumeration string

func (it GenTypeEnumeration) String() string { return genTypeString(it) }
func (it GenTypeEnumeration) kind() string   { return "Enumeration" }
func (it GenTypeEnumeration) key() string    { return "e#" + string(it) }

type GenTypeMapKey string

func (it GenTypeMapKey) String() string { return genTypeString(it) }
func (it GenTypeMapKey) kind() string   { return "MapKey" }
func (it GenTypeMapKey) key() string    { return ":" + string(it) }

type GenTypeReference string

func (it GenTypeReference) String() string { return genTypeString(it) }
func (it GenTypeReference) kind() string   { return "Reference" }
func (it GenTypeReference) key() string    { return "@" + string(it) }

type GenTypeMap struct {
	KeyType   GenType
	ValueType GenType
}

func (it GenTypeMap) String() string { return genTypeString(it) }
func (it GenTypeMap) kind() string   { return "Map" }
func (it GenTypeMap) key() string    { return "[" + it.KeyType.key() + ":" + it.ValueType.key() + "]" }

type GenTypeArray struct {
	ElemType GenType
}

func (it GenTypeArray) String() string { return genTypeString(it) }
func (it GenTypeArray) kind() string   { return "Arr" }
func (it GenTypeArray) key() string    { return "[" + it.ElemType.key() + "]" }

type GenTypeAnd []GenType

func (it GenTypeAnd) String() string { return genTypeString(it) }
func (it GenTypeAnd) kind() string   { return "And" }
func (it GenTypeAnd) key() string {
	return "{" + strings.Join(Map(it, func(gt GenType) string { return gt.key() }), "&") + "}"
}

type GenTypeOr []GenType

func (it GenTypeOr) String() string { return genTypeString(it) }
func (it GenTypeOr) kind() string   { return "Or" }
func (it GenTypeOr) key() string {
	return "{" + strings.Join(Map(it, func(gt GenType) string { return gt.key() }), "|") + "}"
}

type GenTypeTuple []GenType

func (it GenTypeTuple) String() string { return genTypeString(it) }
func (it GenTypeTuple) kind() string   { return "Tup" }
func (it GenTypeTuple) key() string {
	return "{" + strings.Join(Map(it, func(gt GenType) string { return gt.key() }), ",") + "}"
}

type GenTypeLitBool bool

func (it GenTypeLitBool) String() string { return genTypeString(it) }
func (it GenTypeLitBool) kind() string   { return "LitBool" }
func (it GenTypeLitBool) key() string    { return strconv.FormatBool(bool(it)) }

type GenTypeLitString string

func (it GenTypeLitString) String() string { return genTypeString(it) }
func (it GenTypeLitString) kind() string   { return "LitString" }
func (it GenTypeLitString) key() string    { return strconv.Quote(string(it)) }

type GenTypeLitInt int

func (it GenTypeLitInt) String() string { return genTypeString(it) }
func (it GenTypeLitInt) kind() string   { return "LitInt" }
func (it GenTypeLitInt) key() string    { return strconv.FormatInt(int64(it), 36) }

type GenTypeLitStructure []GenTypeLitStructureProperty
type GenTypeLitStructureProperty struct {
	Name     string
	Type     GenType
	Optional bool
}

func (it GenTypeLitStructure) String() string { return genTypeString(it) }
func (it GenTypeLitStructure) kind() string   { return "LitStruct" }
func (it GenTypeLitStructure) key() string {
	return "{" + strings.Join(Map(it, func(p GenTypeLitStructureProperty) string {
		return p.Name + If(p.Optional, "?", "") + " " + p.Type.key()
	}), ";") + "}"
}

func (it *Gen) Type(t GenType) GenType {
	if key := t.key(); it.types[key] == nil {
		it.types[key] = t
	}
	return t
}