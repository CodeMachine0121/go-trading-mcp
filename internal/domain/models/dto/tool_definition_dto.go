package dto

// ToolParameterDto is one box of an ability as the assistant is told about it.
type ToolParameterDto struct {
	Name        string
	Kind        string
	Description string
	IsRequired  bool
}

// ToolDefinitionDto is one ability as the assistant is told about it: what it is
// called, what it does, what has to be filled in, and whether it needs to know who
// is asking.
//
// The last one is on here rather than left implicit because an assistant that cannot
// see it has to discover it by being refused, and being refused is the one outcome
// this connector exists to spare people.
type ToolDefinitionDto struct {
	Name           string
	Description    string
	RequiresSignIn bool
	Parameters     []ToolParameterDto
}
