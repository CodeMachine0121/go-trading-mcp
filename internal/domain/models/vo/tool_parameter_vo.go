package vo

// ToolParameterLocation says where one filled-in box travels once the ask leaves
// this connector: into the address itself, onto the end of it, or inside it.
//
// It exists because the same value means different things in different places: a
// trading symbol names which candle to read when it sits in the address, and names
// which symbol to look for when it sits on the end of it. Only the ability itself
// knows which, so it is declared once, next to the box.
type ToolParameterLocation string

const (
	// ToolParameterInPath puts the value into the address of the thing being named.
	ToolParameterInPath ToolParameterLocation = "path"
	// ToolParameterInQuery puts the value onto the end of the address.
	ToolParameterInQuery ToolParameterLocation = "query"
	// ToolParameterInBody puts the value inside the ask.
	ToolParameterInBody ToolParameterLocation = "body"
)

// ToolParameterKind is the shape one box accepts.
type ToolParameterKind string

const (
	ToolParameterKindString  ToolParameterKind = "string"
	ToolParameterKindInteger ToolParameterKind = "integer"
	ToolParameterKindNumber  ToolParameterKind = "number"
	ToolParameterKindBoolean ToolParameterKind = "boolean"
	ToolParameterKindObject  ToolParameterKind = "object"
	ToolParameterKindArray   ToolParameterKind = "array"
)

// ToolParameterVo is one box an ability asks the assistant to fill in.
//
// The description is not decoration. It is the only thing the assistant reads before
// choosing what to put here, so a box whose description does not say what the
// trading service will refuse is a box the assistant will keep filling in wrongly.
type ToolParameterVo struct {
	Name        string
	Kind        ToolParameterKind
	Description string
	IsRequired  bool
	Location    ToolParameterLocation
}

// NewToolParameterVo declares one box.
func NewToolParameterVo(
	name string,
	kind ToolParameterKind,
	description string,
	isRequired bool,
	location ToolParameterLocation,
) ToolParameterVo {
	return ToolParameterVo{
		Name:        name,
		Kind:        kind,
		Description: description,
		IsRequired:  isRequired,
		Location:    location,
	}
}
