package business

type ModuleRequest map[string]any

type DeleteRequest struct {
	Reason string `json:"reason"`
}

type ModuleActionRequest struct {
	Action string         `json:"action" binding:"required"`
	Data   map[string]any `json:"data"`
}

type ModuleMeta struct {
	Module      string   `json:"module"`
	Title       string   `json:"title"`
	Permission  string   `json:"permission"`
	ListFields  []string `json:"listFields"`
	ActionCodes []string `json:"actionCodes"`
}
