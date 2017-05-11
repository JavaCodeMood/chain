package ivy

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type contract struct {
	name    string
	params  []*param
	clauses []*clause
}

type param struct {
	name string
	typ  typeDesc
}

type clause struct {
	name       string
	params     []*param
	statements []statement

	// decorations
	mintimes, maxtimes []string
}

type statement interface {
	iamaStatement()
}

type verifyStatement struct {
	expr expression

	// Some verify statements are decorated with pointers to associated
	// output statements. Such verifies don't get compiled themselves,
	// but contribute arguments for use in CHECKOUTPUT.
	associatedOutput *outputStatement
}

func (verifyStatement) iamaStatement() {}

type outputStatement struct {
	call *call

	// The AssetAmount expression against which the value is checked
	assetAmount expression

	// Added as a decoration, used by CHECKOUTPUT
	index int64
}

func (outputStatement) iamaStatement() {}

type returnStatement struct {
	expr expression
}

func (returnStatement) iamaStatement() {}

type expression interface {
	String() string
	typ(environ) typeDesc
}

type binaryExpr struct {
	left, right expression
	op          *binaryOp
}

func (e binaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", e.left, e.op.op, e.right)
}

func (e binaryExpr) typ(environ) typeDesc {
	return e.op.result
}

type unaryExpr struct {
	op   *unaryOp
	expr expression
}

func (e unaryExpr) String() string {
	return fmt.Sprintf("%s%s", e.op.op, e.expr)
}

func (e unaryExpr) typ(environ) typeDesc {
	return e.op.result
}

type call struct {
	fn   expression
	args []expression
}

func (e call) String() string {
	var argStrs []string
	for _, a := range e.args {
		argStrs = append(argStrs, a.String())
	}
	return fmt.Sprintf("%s(%s)", e.fn, strings.Join(argStrs, ", "))
}

func (e call) typ(env environ) typeDesc {
	if b := referencedBuiltin(e.fn); b != nil {
		return b.result
	}
	if e.fn.typ(env) == predType {
		return boolType
	}
	return nilType
}

type propRef struct {
	expr     expression
	property string
}

func (p propRef) String() string {
	return fmt.Sprintf("%s.%s", p.expr, p.property)
}

func (e propRef) typ(env environ) typeDesc {
	t := e.expr.typ(env)
	m := properties[t]
	if m != nil {
		return m[e.property]
	}
	return ""
}

type varRef string

func (v varRef) String() string {
	return string(v)
}

func (e varRef) typ(env environ) typeDesc {
	return env[string(e)].t
}

type bytesLiteral []byte

func (e bytesLiteral) String() string {
	return "0x" + hex.EncodeToString([]byte(e))
}

func (bytesLiteral) typ(environ) typeDesc {
	return "String"
}

type integerLiteral int64

func (e integerLiteral) String() string {
	return strconv.FormatInt(int64(e), 10)
}

func (integerLiteral) typ(environ) typeDesc {
	return "Integer"
}

type booleanLiteral bool

func (e booleanLiteral) String() string {
	if e {
		return "true"
	}
	return "false"
}

func (booleanLiteral) typ(environ) typeDesc {
	return "Boolean"
}

type listExpr []expression

func (e listExpr) String() string {
	var elts []string
	for _, elt := range e {
		elts = append(elts, elt.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(elts, ", "))
}

func (listExpr) typ(environ) typeDesc {
	return "List"
}