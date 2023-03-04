package glot

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
)

var tmpls = map[string]*template.Template{}

type GenConfig struct {
	PkgFile           string // eg "go.mod", "package.json" etc
	SrcFileExt        string // eg ".go", ".cs" etc
	CommentLinePrefix string // eg "// ", "# " etc
	FormatCmd         string // eg "go fmt %s" etc
	Named             map[string]string
}

type Gen struct {
	Lang    string
	Version string
	Dot     struct {
		GenConfig
		GenType      string // aka "lsp" or "vsx"
		Items        []any
		FileContents string
	}

	types       map[string]GenType
	dirPathLang string // lang_foo
	dirPathSrc  string // lang_foo/_gen
	dirPathDst  string // lang_foo/GenType_Version
}

type GenDots interface {
	PerEnum(*Gen, func(*GenEnumeration))
}

type GenType interface {
	key() string
	kind() string
	String() string
}

type GenTypeBase string

func (it GenTypeBase) String() string { return Cap(string(it)) }
func (it GenTypeBase) kind() string   { return "Base" }
func (it GenTypeBase) key() string    { return "Base(" + string(it) + ")" }

type GenTypeEnumeration string

func (it GenTypeEnumeration) String() string { return "Enum" + string(it) }
func (it GenTypeEnumeration) kind() string   { return "Enumeration" }
func (it GenTypeEnumeration) key() string    { return "Enumeration(" + string(it) + ")" }

type GenTypeMapKey string

func (it GenTypeMapKey) String() string { return "MapKey" + string(it) }
func (it GenTypeMapKey) kind() string   { return "MapKey" }
func (it GenTypeMapKey) key() string    { return "MapKey(" + string(it) + ")" }

type GenTypeReference string

func (it GenTypeReference) String() string { return string(it) }
func (it GenTypeReference) kind() string   { return "Reference" }
func (it GenTypeReference) key() string    { return "Reference(" + string(it) + ")" }

type GenTypeMap struct {
	KeyType   GenType
	ValueType GenType
}

func (it GenTypeMap) String() string {
	return "Map_" + it.KeyType.String() + "_" + it.ValueType.String()
}
func (it GenTypeMap) kind() string { return "Map" }
func (it GenTypeMap) key() string  { return "Map[" + it.KeyType.key() + "]" + it.ValueType.key() }

type GenTypeArr struct {
	ElemType GenType
}

func (it GenTypeArr) String() string { return "Arr_" + it.ElemType.String() }
func (it GenTypeArr) kind() string   { return "Arr" }
func (it GenTypeArr) key() string    { return "[" + it.ElemType.key() + "]" }

type GenTypeAnd []GenType

func (it GenTypeAnd) String() string {
	return "And_" + strings.Join(Map(it, func(gt GenType) string { return gt.String() }), "_")
}
func (it GenTypeAnd) kind() string { return "And" }
func (it GenTypeAnd) key() string {
	return "And(" + strings.Join(Map(it, func(gt GenType) string { return gt.key() }), ",") + ")"
}

type GenTypeOr []GenType

func (it GenTypeOr) String() string {
	return "Or_" + strings.Join(Map(it, func(gt GenType) string { return gt.String() }), "_")
}
func (it GenTypeOr) kind() string { return "Or" }
func (it GenTypeOr) key() string {
	return "Or(" + strings.Join(Map(it, func(gt GenType) string { return gt.key() }), ",") + ")"
}

type GenTypeTup []GenType

func (it GenTypeTup) String() string {
	return "Tup_" + strings.Join(Map(it, func(gt GenType) string { return gt.String() }), "_")
}
func (it GenTypeTup) kind() string { return "Tup" }
func (it GenTypeTup) key() string {
	return "Tup(" + strings.Join(Map(it, func(gt GenType) string { return gt.key() }), ",") + ")"
}

type GenTypeLitBool bool

func (it GenTypeLitBool) String() string { return "Bool_" + strconv.FormatBool(bool(it)) }
func (it GenTypeLitBool) kind() string   { return "LitBool" }
func (it GenTypeLitBool) key() string    { return "LitBool(" + strconv.FormatBool(bool(it)) + ")" }

type GenTypeLitString string

func (it GenTypeLitString) String() string { return "String_" + string(it) }
func (it GenTypeLitString) kind() string   { return "LitString" }
func (it GenTypeLitString) key() string    { return "LitString(" + strconv.Quote(string(it)) + ")" }

type GenTypeLitInt int

func (it GenTypeLitInt) String() string { return "Int_" + strconv.Itoa(int(it)) }
func (it GenTypeLitInt) kind() string   { return "LitInt" }
func (it GenTypeLitInt) key() string    { return "LitInt(" + strconv.FormatInt(int64(it), 36) + ")" }

type GenTypeLitStruct []GenTypeLitStructProperty

type GenTypeLitStructProperty struct {
	Name     string
	Type     GenType
	Optional bool
}

func (it GenTypeLitStruct) String() string { panic(">>>GenTypeLitStruct") }
func (it GenTypeLitStruct) kind() string   { return "LitStruct" }
func (it GenTypeLitStruct) key() string    { panic("GenTypeLitStruct:") }

type GenBase struct {
	DocLines   []string
	Since      string
	Deprecated string
}

func (it *GenBase) base() *GenBase { return it }

type GenEnumeration struct {
	GenBase
	Name       string
	Type       GenType
	Enumerants []GenEnumerant
}
type GenEnumerant struct {
	GenBase
	Name    string
	NameCap string
	Value   string
}

func (it *Gen) tmpl(filePath string) (ret *template.Template) {
	if ret = tmpls[filePath]; ret == nil {
		ret = template.Must(template.ParseFiles(filePath))
		tmpls[filePath] = ret
	}
	return
}

func (it *Gen) Type(t GenType) GenType {
	if key := t.key(); it.types[key] == nil {
		it.types[key] = t
	}
	return t
}

func (it *Gen) Generate(dots GenDots) {
	it.types = map[string]GenType{}
	it.dirPathLang = "lang_" + it.Lang
	it.Dot.GenConfig = LoadFromJSONFile[GenConfig](it.dirPathLang + "/" + it.dirPathLang + ".json")
	it.dirPathSrc = "lang_" + it.Lang + "/_gen"
	it.dirPathDst = it.dirPathLang + "/" + it.Dot.GenType + "_" + it.Version
	println(it.dirPathDst + "...")
	DirCreate(it.dirPathDst)

	var buf bytes.Buffer
	for file_name, dots := range map[string]func() []any{
		"types_enums": toAnys(it, dots.PerEnum),
	} {
		tmpl := it.tmpl(it.dirPathSrc + "/" + file_name + ".tmpl")
		it.Dot.Items = dots()
		it.tmplExec(&buf, tmpl)

		it.toCodeFile(&buf, file_name)
	}
	{
		types := make([]any, 0, len(it.types))
		for _, ty := range it.types {
			tmpl := it.tmpl(it.dirPathSrc + "/type_" + ty.kind() + ".tmpl")
			it.Dot.Items = []any{ty}
			it.tmplExec(&buf, tmpl)
			types = append(types, it.Dot.FileContents)
		}
		tmpl := it.tmpl(it.dirPathSrc + "/types.tmpl")
		it.Dot.Items = types
		it.tmplExec(&buf, tmpl)
		it.toCodeFile(&buf, "types")
	}
}

func (it *Gen) format(filePath string) {
	if it.Dot.FormatCmd != "" {
		parts := Replace(Without(strings.Split(it.Dot.FormatCmd, " "), ""), "%s", filePath)
		_ = exec.Command(parts[0], parts[1:]...).Run()
	}
}

func (it *Gen) toCodeFile(buf *bytes.Buffer, fileName string) {
	tmpl := it.tmpl(it.dirPathSrc + "/" + "file_code.tmpl")
	it.tmplExec(buf, tmpl)
	file_path := it.dirPathDst + "/" + fileName + it.Dot.SrcFileExt
	FileWrite(file_path, []byte(it.Dot.FileContents))
	it.format(file_path)
}

func (it *Gen) tmplExec(buf *bytes.Buffer, tmpl *template.Template) {
	buf.Reset()
	if err := tmpl.Execute(buf, &it.Dot); err != nil {
		panic(err)
	}
	it.Dot.FileContents = buf.String()
}

func toAnys[T any](gen *Gen, each func(*Gen, func(*T))) func() []any {
	return func() (ret []any) {
		each(gen, func(dot *T) {
			ret = append(ret, dot)
		})
		return
	}
}
