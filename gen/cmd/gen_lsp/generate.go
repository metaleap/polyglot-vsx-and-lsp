package main

import (
	"strings"

	glot "polyglot-vsx-and-lsp/pkg"
)

func generate(metaModel *MetaModel, ver string, lang string) {
	gen := glot.Gen{LangIdent: lang, Dot: glot.GenDot{
		GenTitle: "Language Server Protocol (LSP)",
		GenIdent: "lsp",
		GenVer:   ver,
		GenRepo:  "github.com/metaleap/polyglot-vsx-and-lsp/gen/cmd/gen_lsp",
	}}
	gen.Generate(metaModel)
}

func (*MetaModel) toGenBase(it *MMBase) glot.GenBase {
	it.Documentation = strings.TrimSpace(it.Documentation)

	return glot.GenBase{Deprecated: it.Deprecated, Since: it.Since, Name: it.Name, NameUp: glot.Up0(it.Name),
		DocLines: glot.If[[]string](it.Documentation == "", nil, strings.Split(it.Documentation, "\n"))}
}

func (it *MetaModel) PerEnumeration(gen *glot.Gen, do func(*glot.GenEnumeration)) {
	for _, enumeration := range it.Enumerations {
		if enumeration.Proposed {
			continue
		}
		do(&glot.GenEnumeration{
			GenBase: it.toGenBase(&enumeration.MMBase),
			Type:    gen.Type(enumeration.Type.toGenType()),
			Enumerants: glot.Map(glot.Filter(enumeration.Values, func(e MMEnumerant) bool { return !e.Proposed }), func(e MMEnumerant) glot.GenEnumerant {
				return glot.GenEnumerant{
					GenBase: it.toGenBase(&e.MMBase),
					Value:   e.Value.String(),
				}
			}),
		})
	}
}

func (it *MetaModel) PerStructure(gen *glot.Gen, do func(*glot.GenStructure)) {
	for _, structure := range it.Structures {
		if structure.Proposed {
			continue
		}
		do(&glot.GenStructure{
			GenBase: it.toGenBase(&structure.MMBase),
			Mixins:  glot.Map(structure.Mixins, func(t MMType) glot.GenType { return gen.Type(t.toGenType()) }),
			Extends: glot.Map(structure.Extends, func(t MMType) glot.GenType { return gen.Type(t.toGenType()) }),
			Properties: glot.Map(glot.Filter(structure.Properties, func(p MMProperty) bool { return !p.Proposed }), func(p MMProperty) glot.GenProperty {
				return glot.GenProperty{
					GenBase:  it.toGenBase(&p.MMBase),
					Type:     gen.Type(p.Type.toGenType()),
					Optional: p.Optional,
				}
			}),
		})
	}
}

func (it *MMType) toGenType() glot.GenType {
	switch it.Kind {
	case MMTypeKindBase:
		return glot.GenTypeBaseType(it.Name)
	case MMTypeKindReference:
		return glot.GenTypeReference(it.Name)
	case MMTypeKindArray:
		return glot.GenTypeArray{it.Element.toGenType()}
	case MMTypeKindMap:
		return glot.GenTypeMap{KeyType: it.Key.toGenType(), ValueType: it.Value.t.toGenType()}
	case MMTypeKindAnd:
		return glot.GenTypeAnd(glot.Map(it.Items, func(t MMType) glot.GenType { return t.toGenType() }))
	case MMTypeKindOr:
		return glot.GenTypeOr(glot.Map(it.Items, func(t MMType) glot.GenType { return t.toGenType() }))
	case MMTypeKindTuple:
		return glot.GenTypeTuple(glot.Map(it.Items, func(t MMType) glot.GenType { return t.toGenType() }))
	case MMTypeKindStringLiteral:
		return glot.GenTypeLitString(it.Value.s)
	case MMTypeKindIntegerLiteral:
		return glot.GenTypeLitInt(it.Value.i)
	case MMTypeKindBooleanLiteral:
		return glot.GenTypeLitBool(it.Value.b)
	case MMTypeKindLiteral:
		return glot.GenTypeLitStructure(glot.Map(it.Value.l.Properties, func(p MMProperty) glot.GenTypeLitStructureProperty {
			return glot.GenTypeLitStructureProperty{Name: p.Name, Optional: p.Optional, Type: p.Type.toGenType()}
		}))
	}
	panic(it.Kind)
}
