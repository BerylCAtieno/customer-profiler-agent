package models

// CustomerProfile represents a detailed customer persona
type CustomerProfile struct {
	Age               string   `json:"age"`
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

// ProfileResponse contains mulriple customer profiles related to a given business idea

type ProfileResponse struct {
	BusinessIdea string            `json:"business_idea"`
	Profiles     []CustomerProfile `json:"profiles"`
	Summary      string            `json:"summary"`
	Keywords     []string          `json:"keywords"`
}
