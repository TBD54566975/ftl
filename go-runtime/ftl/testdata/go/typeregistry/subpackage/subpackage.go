package subpackage

//ftl:enum
type StringsTypeEnum interface {
	tag()
}

type Single string

func (Single) tag() {}

type List []string

func (List) tag() {}

type Object struct {
	S string
}

func (Object) tag() {}
