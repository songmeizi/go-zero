package ast

import (
	"fmt"
	"sort"

	"github.com/tal-tech/go-zero/tools/goctl/api/parser/g4/gen/api"
	"github.com/tal-tech/go-zero/tools/goctl/api/util"
)

type (
	// TypeExpr describes an expression for TypeAlias and TypeStruct
	TypeExpr interface {
		Doc() []Expr
		Format() error
		Equal(v interface{}) bool
		NameExpr() Expr
	}

	// TypeAlias describes alias ast for api syntax
	TypeAlias struct {
		Name        Expr
		Assign      Expr
		DataType    DataType
		DocExpr     []Expr
		CommentExpr Expr
	}

	// TypeStruct describes structure ast for api syntax
	TypeStruct struct {
		Name    Expr
		Struct  Expr
		LBrace  Expr
		RBrace  Expr
		DocExpr []Expr
		Fields  []*TypeField
	}

	// TypeField describes field ast for api syntax
	TypeField struct {
		IsAnonymous bool
		// Name is nil if IsAnonymous
		Name        Expr
		DataType    DataType
		Tag         Expr
		DocExpr     []Expr
		CommentExpr Expr
	}

	// DataType describes datatype for api syntax, the default implementation expressions are
	// Literal, Interface, Map, Array, Time, Pointer
	DataType interface {
		Expr() Expr
		Equal(dt DataType) bool
		Format() error
		IsNotNil() bool
	}

	// Literal describes the basic types of golang, non-reference types,
	// such as int, bool, Foo,foo.Bar...
	Literal struct {
		Package *Package
		Literal Expr
	}

	// Interface describes the interface type of golang,Its fixed value is interface{}
	Interface struct {
		Literal Expr
	}

	// Map describes the map ast for api syntax
	Map struct {
		MapExpr Expr
		Map     Expr
		LBrack  Expr
		RBrack  Expr
		Key     Expr
		Value   DataType
	}

	// Array describes the slice ast for api syntax
	Array struct {
		ArrayExpr Expr
		LBrack    Expr
		RBrack    Expr
		Literal   DataType
	}

	// Time describes the time ast for api syntax
	Time struct {
		Literal Expr
	}

	// Pointer describes the pointer ast for api syntax
	Pointer struct {
		Package     *Package
		PointerExpr Expr
		Star        Expr
		Name        Expr
	}

	// Package describes the package of type
	Package struct {
		Name Expr
		Dot  Expr
	}
)

// VisitTypeSpec implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitTypeSpec(ctx *api.TypeSpecContext) interface{} {
	if ctx.TypeLit() != nil {
		return []TypeExpr{ctx.TypeLit().Accept(v).(TypeExpr)}
	}
	return ctx.TypeBlock().Accept(v)
}

// VisitTypeLit implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitTypeLit(ctx *api.TypeLitContext) interface{} {
	typeLit := ctx.TypeLitBody().Accept(v)
	alias, ok := typeLit.(*TypeAlias)
	if ok {
		return alias
	}

	st, ok := typeLit.(*TypeStruct)
	if ok {
		return st
	}

	return typeLit
}

// VisitTypeBlock implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitTypeBlock(ctx *api.TypeBlockContext) interface{} {
	list := ctx.AllTypeBlockBody()
	var types []TypeExpr
	for _, each := range list {
		types = append(types, each.Accept(v).(TypeExpr))

	}
	return types
}

// VisitTypeLitBody implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitTypeLitBody(ctx *api.TypeLitBodyContext) interface{} {
	if ctx.TypeAlias() != nil {
		return ctx.TypeAlias().Accept(v)
	}
	return ctx.TypeStruct().Accept(v)
}

// VisitTypeBlockBody implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitTypeBlockBody(ctx *api.TypeBlockBodyContext) interface{} {
	if ctx.TypeBlockAlias() != nil {
		return ctx.TypeBlockAlias().Accept(v).(*TypeAlias)
	}
	return ctx.TypeBlockStruct().Accept(v).(*TypeStruct)
}

// VisitTypeStruct implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitTypeStruct(ctx *api.TypeStructContext) interface{} {
	var st TypeStruct
	st.Name = v.newExprWithToken(ctx.GetStructName())

	if util.UnExport(ctx.GetStructName().GetText()) {

	}
	if ctx.GetStructToken() != nil {
		structExpr := v.newExprWithToken(ctx.GetStructToken())
		structTokenText := ctx.GetStructToken().GetText()
		if structTokenText != "struct" {
			v.panic(structExpr, fmt.Sprintf("expecting 'struct', found input '%s'", structTokenText))
		}

		if api.IsGolangKeyWord(structTokenText, "struct") {
			v.panic(structExpr, fmt.Sprintf("expecting 'struct', but found golang keyword '%s'", structTokenText))
		}

		st.Struct = structExpr
	}

	st.LBrace = v.newExprWithToken(ctx.GetLbrace())
	st.RBrace = v.newExprWithToken(ctx.GetRbrace())
	fields := ctx.AllField()
	for _, each := range fields {
		f := each.Accept(v)
		if f == nil {
			continue
		}
		st.Fields = append(st.Fields, f.(*TypeField))
	}
	return &st
}

// VisitTypeBlockStruct implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitTypeBlockStruct(ctx *api.TypeBlockStructContext) interface{} {
	var st TypeStruct
	st.Name = v.newExprWithToken(ctx.GetStructName())

	if ctx.GetStructToken() != nil {
		structExpr := v.newExprWithToken(ctx.GetStructToken())
		structTokenText := ctx.GetStructToken().GetText()
		if structTokenText != "struct" {
			v.panic(structExpr, fmt.Sprintf("expecting 'struct', found imput '%s'", structTokenText))
		}

		if api.IsGolangKeyWord(structTokenText, "struct") {
			v.panic(structExpr, fmt.Sprintf("expecting 'struct', but found golang keyword '%s'", structTokenText))
		}

		st.Struct = structExpr
	}
	st.DocExpr = v.getDoc(ctx)
	st.LBrace = v.newExprWithToken(ctx.GetLbrace())
	st.RBrace = v.newExprWithToken(ctx.GetRbrace())
	fields := ctx.AllField()
	for _, each := range fields {
		f := each.Accept(v)
		if f == nil {
			continue
		}
		st.Fields = append(st.Fields, f.(*TypeField))
	}
	return &st
}

// VisitTypeBlockAlias implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitTypeBlockAlias(ctx *api.TypeBlockAliasContext) interface{} {
	var alias TypeAlias
	alias.Name = v.newExprWithToken(ctx.GetAlias())
	alias.Assign = v.newExprWithToken(ctx.GetAssign())
	alias.DataType = ctx.DataType().Accept(v).(DataType)
	alias.DocExpr = v.getDoc(ctx)
	alias.CommentExpr = v.getComment(ctx)
	// todo: reopen if necessary
	v.panic(alias.Name, "unsupported alias")
	return &alias
}

// VisitTypeAlias implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitTypeAlias(ctx *api.TypeAliasContext) interface{} {
	var alias TypeAlias
	alias.Name = v.newExprWithToken(ctx.GetAlias())
	alias.Assign = v.newExprWithToken(ctx.GetAssign())
	alias.DataType = ctx.DataType().Accept(v).(DataType)
	alias.DocExpr = v.getDoc(ctx)
	alias.CommentExpr = v.getComment(ctx)
	// todo: reopen if necessary
	v.panic(alias.Name, "unsupported alias")
	return &alias
}

// VisitField implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitField(ctx *api.FieldContext) interface{} {
	iAnonymousFiled := ctx.AnonymousFiled()
	iNormalFieldContext := ctx.NormalField()
	if iAnonymousFiled != nil {
		return iAnonymousFiled.Accept(v).(*TypeField)
	}
	if iNormalFieldContext != nil {
		return iNormalFieldContext.Accept(v).(*TypeField)
	}
	return nil
}

// VisitNormalField implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitNormalField(ctx *api.NormalFieldContext) interface{} {
	var field TypeField
	field.Name = v.newExprWithToken(ctx.GetFieldName())

	iDataTypeContext := ctx.DataType()
	if iDataTypeContext != nil {
		field.DataType = iDataTypeContext.Accept(v).(DataType)
		field.CommentExpr = v.getComment(ctx)
	}
	if ctx.GetTag() != nil {
		tagText := ctx.GetTag().GetText()
		tagExpr := v.newExprWithToken(ctx.GetTag())
		if !api.MatchTag(tagText) {
			v.panic(tagExpr, fmt.Sprintf("mismatched tag, found input '%s'", tagText))
		}
		field.Tag = tagExpr
		field.CommentExpr = v.getComment(ctx)
	}
	field.DocExpr = v.getDoc(ctx)
	return &field
}

// VisitAnonymousFiled implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitAnonymousFiled(ctx *api.AnonymousFiledContext) interface{} {
	start := ctx.GetStart()
	stop := ctx.GetStop()
	var field TypeField
	field.IsAnonymous = true
	var pkg *Package
	if ctx.PackageExpr() != nil {
		p := ctx.PackageExpr().Accept(v)
		pkg = p.(*Package)
	}
	if ctx.GetStar() != nil {
		nameExpr := v.newExprWithTerminalNode(ctx.ID())
		pointerExpr := v.newExprWithText(ctx.GetStar().GetText()+ctx.ID().GetText(), start.GetLine(), start.GetColumn(), start.GetStart(), stop.GetStop())
		if pkg != nil {
			pointerExpr = v.newExprWithText(ctx.GetStar().GetText()+pkg.Name.Text()+"."+ctx.ID().GetText(), start.GetLine(), start.GetColumn(), start.GetStart(), stop.GetStop())
		}
		field.DataType = &Pointer{
			Package:     pkg,
			PointerExpr: pointerExpr,
			Star:        v.newExprWithToken(ctx.GetStar()),
			Name:        nameExpr,
		}
	} else {
		nameExpr := v.newExprWithTerminalNode(ctx.ID())
		if pkg != nil {
			nameExpr = v.newExprWithText(pkg.Name.Text()+"."+ctx.ID().GetText(), start.GetLine(), start.GetColumn(), start.GetStart(), stop.GetStop())
		}
		field.DataType = &Literal{
			Package: pkg,
			Literal: nameExpr,
		}
	}

	field.DocExpr = v.getDoc(ctx)
	field.CommentExpr = v.getComment(ctx)
	return &field
}

// VisitDataType implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitDataType(ctx *api.DataTypeContext) interface{} {
	if ctx.ID() != nil {
		var pkg *Package
		if ctx.PackageExpr() != nil {
			p := ctx.PackageExpr().Accept(v)
			pkg = p.(*Package)
		}

		idExpr := v.newExprWithTerminalNode(ctx.ID())
		if pkg != nil {
			idExpr = v.newExprWithText(pkg.Name.Text()+"."+idExpr.Text(), pkg.Name.Line(), pkg.Name.Column(), pkg.Name.Start(), idExpr.Stop())
		}
		return &Literal{Package: pkg, Literal: idExpr}
	}

	if ctx.MapType() != nil {
		t := ctx.MapType().Accept(v)
		return t
	}

	if ctx.ArrayType() != nil {
		return ctx.ArrayType().Accept(v)
	}

	if ctx.GetInter() != nil {
		return &Interface{Literal: v.newExprWithToken(ctx.GetInter())}
	}

	if ctx.GetTime() != nil {
		// todo: reopen if it is necessary
		timeExpr := v.newExprWithToken(ctx.GetTime())
		v.panic(timeExpr, "unsupported time.Time")
		return &Time{Literal: timeExpr}
	}

	if ctx.PointerType() != nil {
		return ctx.PointerType().Accept(v)
	}

	return ctx.TypeStruct().Accept(v)
}

// VisitPointerType implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitPointerType(ctx *api.PointerTypeContext) interface{} {
	nameExpr := v.newExprWithTerminalNode(ctx.ID())
	var pkg *Package
	if ctx.PackageExpr() != nil {
		p := ctx.PackageExpr().Accept(v)
		pkg = p.(*Package)
	}

	return &Pointer{
		Package:     pkg,
		PointerExpr: v.newExprWithText(ctx.GetText(), ctx.GetStar().GetLine(), ctx.GetStar().GetColumn(), ctx.GetStar().GetStart(), ctx.ID().GetSymbol().GetStop()),
		Star:        v.newExprWithToken(ctx.GetStar()),
		Name:        nameExpr,
	}
}

// VisitPackageExpr implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitPackageExpr(ctx *api.PackageExprContext) interface{} {
	nameExpr := v.newExprWithToken(ctx.GetPackageName())
	return &Package{
		Name: nameExpr,
		Dot:  v.newExprWithToken(ctx.GetDot()),
	}
}

// VisitMapType implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitMapType(ctx *api.MapTypeContext) interface{} {
	return &Map{
		MapExpr: v.newExprWithText(ctx.GetText(), ctx.GetMapToken().GetLine(), ctx.GetMapToken().GetColumn(),
			ctx.GetMapToken().GetStart(), ctx.GetValue().GetStop().GetStop()),
		Map:    v.newExprWithToken(ctx.GetMapToken()),
		LBrack: v.newExprWithToken(ctx.GetLbrack()),
		RBrack: v.newExprWithToken(ctx.GetRbrack()),
		Key:    v.newExprWithToken(ctx.GetKey()),
		Value:  ctx.GetValue().Accept(v).(DataType),
	}
}

// VisitArrayType implements from api.BaseApiParserVisitor
func (v *ApiVisitor) VisitArrayType(ctx *api.ArrayTypeContext) interface{} {
	return &Array{
		ArrayExpr: v.newExprWithText(ctx.GetText(), ctx.GetLbrack().GetLine(), ctx.GetLbrack().GetColumn(), ctx.GetLbrack().GetStart(), ctx.DataType().GetStop().GetStop()),
		LBrack:    v.newExprWithToken(ctx.GetLbrack()),
		RBrack:    v.newExprWithToken(ctx.GetRbrack()),
		Literal:   ctx.DataType().Accept(v).(DataType),
	}
}

// NameExpr returns the expression string of TypeAlias
func (a *TypeAlias) NameExpr() Expr {
	return a.Name
}

// Doc returns the document of TypeAlias, like // some text
func (a *TypeAlias) Doc() []Expr {
	return a.DocExpr
}

// Comment returns the comment of TypeAlias, like // some text
func (a *TypeAlias) Comment() Expr {
	return a.CommentExpr
}

// Format provides a formatter for api command, now nothing to do
func (a *TypeAlias) Format() error {
	return nil
}

// Equal compares whether the element literals in two TypeAlias are equal
func (a *TypeAlias) Equal(v interface{}) bool {
	if v == nil {
		return false
	}

	alias := v.(*TypeAlias)
	if !a.Name.Equal(alias.Name) {
		return false
	}

	if !a.Assign.Equal(alias.Assign) {
		return false
	}

	if !a.DataType.Equal(alias.DataType) {
		return false
	}

	return EqualDoc(a, alias)
}

// Expr returns the expression string of Literal
func (l *Literal) Expr() Expr {
	return l.Literal
}

// Format provides a formatter for api command, now nothing to do
func (l *Literal) Format() error {
	// todo
	return nil
}

// Equal compares whether the element literals in two Literal are equal
func (l *Literal) Equal(dt DataType) bool {
	if dt == nil {
		return false
	}

	v, ok := dt.(*Literal)
	if !ok {
		return false
	}

	if l.Package != nil {
		if !l.Package.Equal(v.Package) {
			return false
		}
	}

	return l.Literal.Equal(v.Literal)
}

// IsNotNil returns whether the instance is nil or not
func (l *Literal) IsNotNil() bool {
	return l != nil
}

// Expr returns the expression string of Interface
func (i *Interface) Expr() Expr {
	return i.Literal
}

// Format provides a formatter for api command, now nothing to do
func (i *Interface) Format() error {
	// todo
	return nil
}

// Equal compares whether the element literals in two Interface are equal
func (i *Interface) Equal(dt DataType) bool {
	if dt == nil {
		return false
	}

	v, ok := dt.(*Interface)
	if !ok {
		return false
	}

	return i.Literal.Equal(v.Literal)
}

// IsNotNil returns whether the instance is nil or not
func (i *Interface) IsNotNil() bool {
	return i != nil
}

// Expr returns the expression string of Map
func (m *Map) Expr() Expr {
	return m.MapExpr
}

// Format provides a formatter for api command, now nothing to do
func (m *Map) Format() error {
	// todo
	return nil
}

// Equal compares whether the element literals in two Map are equal
func (m *Map) Equal(dt DataType) bool {
	if dt == nil {
		return false
	}

	v, ok := dt.(*Map)
	if !ok {
		return false
	}

	if !m.Key.Equal(v.Key) {
		return false
	}

	if !m.Value.Equal(v.Value) {
		return false
	}

	if !m.MapExpr.Equal(v.MapExpr) {
		return false
	}

	return m.Map.Equal(v.Map)
}

// IsNotNil returns whether the instance is nil or not
func (m *Map) IsNotNil() bool {
	return m != nil
}

// Expr returns the expression string of Array
func (a *Array) Expr() Expr {
	return a.ArrayExpr
}

// Format provides a formatter for api command, now nothing to do
func (a *Array) Format() error {
	// todo
	return nil
}

// Equal compares whether the element literals in two Array are equal
func (a *Array) Equal(dt DataType) bool {
	if dt == nil {
		return false
	}

	v, ok := dt.(*Array)
	if !ok {
		return false
	}

	if !a.ArrayExpr.Equal(v.ArrayExpr) {
		return false
	}

	return a.Literal.Equal(v.Literal)
}

// IsNotNil returns whether the instance is nil or not
func (a *Array) IsNotNil() bool {
	return a != nil
}

// Expr returns the expression string of Time
func (t *Time) Expr() Expr {
	return t.Literal
}

// Format provides a formatter for api command, now nothing to do
func (t *Time) Format() error {
	// todo
	return nil
}

// Equal compares whether the element literals in two Time are equal
func (t *Time) Equal(dt DataType) bool {
	if dt == nil {
		return false
	}

	v, ok := dt.(*Time)
	if !ok {
		return false
	}

	return t.Literal.Equal(v.Literal)
}

// IsNotNil returns whether the instance is nil or not
func (t *Time) IsNotNil() bool {
	return t != nil
}

// Expr returns the expression string of Pointer
func (p *Pointer) Expr() Expr {
	return p.PointerExpr
}

// Format provides a formatter for api command, now nothing to do
func (p *Pointer) Format() error {
	return nil
}

// Equal compares whether the element literals in two Pointer are equal
func (p *Pointer) Equal(dt DataType) bool {
	if dt == nil {
		return false
	}

	v, ok := dt.(*Pointer)
	if !ok {
		return false
	}

	if !p.PointerExpr.Equal(v.PointerExpr) {
		return false
	}

	if !p.Star.Equal(v.Star) {
		return false
	}

	if p.Package != nil {
		if !p.Package.Equal(v.Package) {
			return false
		}
	}

	return p.Name.Equal(v.Name)
}

// IsNotNil returns whether the instance is nil or not
func (p *Pointer) IsNotNil() bool {
	return p != nil
}

// Format provides a formatter for api command, now nothing to do
func (p *Package) Format() error {
	return nil
}

// Equal compares whether the element literals in two Package are equal
func (p *Package) Equal(pkg *Package) bool {
	if pkg == nil {
		return false
	}

	return p.Name.Equal(pkg.Name)
}

// IsNotNil returns whether the instance is nil or not
func (p *Package) IsNotNil() bool {
	return p != nil
}

// NameExpr returns the expression string of TypeStruct
func (s *TypeStruct) NameExpr() Expr {
	return s.Name
}

// Equal compares whether the element literals in two TypeStruct are equal
func (s *TypeStruct) Equal(dt interface{}) bool {
	if dt == nil {
		return false
	}

	v, ok := dt.(*TypeStruct)
	if !ok {
		return false
	}

	if !s.Name.Equal(v.Name) {
		return false
	}

	var expectDoc, actualDoc []Expr
	expectDoc = append(expectDoc, s.DocExpr...)
	actualDoc = append(actualDoc, v.DocExpr...)
	sort.Slice(expectDoc, func(i, j int) bool {
		return expectDoc[i].Line() < expectDoc[j].Line()
	})

	for index, each := range actualDoc {
		if !each.Equal(actualDoc[index]) {
			return false
		}
	}

	if s.Struct != nil {
		if s.Struct != nil {
			if !s.Struct.Equal(v.Struct) {
				return false
			}
		}
	}

	if len(s.Fields) != len(v.Fields) {
		return false
	}

	var expected, acual []*TypeField
	expected = append(expected, s.Fields...)
	acual = append(acual, v.Fields...)

	sort.Slice(expected, func(i, j int) bool {
		return expected[i].DataType.Expr().Line() < expected[j].DataType.Expr().Line()
	})
	sort.Slice(acual, func(i, j int) bool {
		return acual[i].DataType.Expr().Line() < acual[j].DataType.Expr().Line()
	})

	for index, each := range expected {
		ac := acual[index]
		if !each.Equal(ac) {
			return false
		}
	}

	return true
}

// Doc returns the document of TypeStruct, like // some text
func (s *TypeStruct) Doc() []Expr {
	return s.DocExpr
}

// Format provides a formatter for api command, now nothing to do
func (s *TypeStruct) Format() error {
	// todo
	return nil
}

// Equal compares whether the element literals in two TypeField are equal
func (t *TypeField) Equal(v interface{}) bool {
	if v == nil {
		return false
	}

	f, ok := v.(*TypeField)
	if !ok {
		return false
	}

	if t.IsAnonymous != f.IsAnonymous {
		return false
	}

	if !t.DataType.Equal(f.DataType) {
		return false
	}

	if !t.IsAnonymous {
		if !t.Name.Equal(f.Name) {
			return false
		}

		if t.Tag != nil {
			if !t.Tag.Equal(f.Tag) {
				return false
			}
		}
	}

	return EqualDoc(t, f)
}

// Doc returns the document of TypeField, like // some text
func (t *TypeField) Doc() []Expr {
	return t.DocExpr
}

// Comment returns the comment of TypeField, like // some text
func (t *TypeField) Comment() Expr {
	return t.CommentExpr
}

// Format provides a formatter for api command, now nothing to do
func (t *TypeField) Format() error {
	// todo
	return nil
}
