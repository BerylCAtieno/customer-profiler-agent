package models

type CustomerProfile struct {
	Age               int      `json:"age"`
	Gender            string   `json:"gender"`
	Location          string   `json:"location"`
	Occupation        string   `json:"occupation"`
	Income            string   `json:"income"`
	Motivations       []string `json:"motivations"`
	Interests         []string `json:"interests"`
	PainPoints        []string `json:"pain_points"`
	BuyingBehaviors   []string `json:"buying_behaviors"`
	PreferredChannels []string `json:"preferred_channels"`
}
