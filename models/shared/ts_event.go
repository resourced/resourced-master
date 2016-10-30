package shared

type TSEventHighchartLinePayload struct {
	ID          int64  `json:"ID"`
	CreatedFrom int64  `json:"CreatedFrom"`
	CreatedTo   int64  `json:"CreatedTo"`
	Description string `json:"Description"`
}

type TSEventCreatePayload struct {
	From        int64  `json:"from"`
	To          int64  `json:"to"`
	Description string `json:"description"`
}
