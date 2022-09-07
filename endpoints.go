package playfab

// Filter represents a filter for the PlayFab catalog search system.
type Filter struct {
	Count   bool   `json:"count"`
	Filter  string `json:"filter"`
	OrderBy string `json:"orderBy"`
	SCID    string `json:"scid"`
	Skip    int    `json:"skip,omitempty"`
	Limit   int    `json:"top"`
}

// Search searches the PlayFab catalog for items matching the given filter.
func (p *PlayFab) Search(filter Filter) (map[string]any, error) {
	m := make(map[string]any)
	if err := p.request("Catalog/Search", filter, &m); err != nil {
		return nil, err
	}
	return m["data"].(map[string]any), nil
}

// TODO: Implement more endpoints.
