package lcd

type StakingPoolResponse struct {
	Pool struct {
		BondedTokens string `json:"bonded_tokens"`
		NotBonded    string `json:"not_bonded_tokens"`
	} `json:"pool"`
}
