package stake

// 参考文档：https://prv.docs.biya.io/api-reference/stake/get-validators

type GetValidatorsResponse struct {
	Validators []Validator `json:"validators"`
	Pagination struct {
		Page       int    `json:"page"`
		PageSize   int    `json:"pageSize"`
		Total      string `json:"total"`
		TotalPages int    `json:"totalPages"`
		HasPrev    bool   `json:"hasPrev"`
		HasNext    bool   `json:"hasNext"`
	} `json:"pagination"`
}

type Validator struct {
	ID               string  `json:"id"`
	Moniker          string  `json:"moniker"`
	OperatorAddress  string  `json:"operatorAddress"`
	ConsensusAddress string  `json:"consensusAddress"`
	Jailed           bool    `json:"jailed"`
	Status           int     `json:"status"`
	Tokens           string  `json:"tokens"`
	UptimePercentage float64 `json:"uptimePercentage"`
}
