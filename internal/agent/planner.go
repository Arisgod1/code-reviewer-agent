package agent

func NextAction(s *State) string {
	if s.Done {
		return "finish"
	}
	if s.ParsedLines == nil {
		return "parse_diff"
	}
	if s.Findings == nil {
		return "scan_risk_rules"
	}
	return "finish"
}
