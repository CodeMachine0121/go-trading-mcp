package dto

// ToolParameterDto is one box of an ability as the assistant is told about it.
type ToolParameterDto struct {
	Name        string
	Kind        string
	Description string
	IsRequired  bool
}

// ToolDefinitionDto is one ability as the assistant is told about it: what it is
// called, what it does, and what has to be filled in.
type ToolDefinitionDto struct {
	Name        string
	Description string
	Parameters  []ToolParameterDto
}
