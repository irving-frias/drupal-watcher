//go:build worldcup

package worldcup

type Game struct {
	ID              string `json:"id"`
	HomeTeamID      string `json:"home_team_id"`
	AwayTeamID      string `json:"away_team_id"`
	HomeScore       string `json:"home_score"`
	AwayScore       string `json:"away_score"`
	HomeScorers     string `json:"home_scorers"`
	AwayScorers     string `json:"away_scorers"`
	Group           string `json:"group"`
	Matchday        string `json:"matchday"`
	LocalDate       string `json:"local_date"`
	Finished        string `json:"finished"`
	TimeElapsed     string `json:"time_elapsed"`
	Type            string `json:"type"`
	HomeTeamNameEn  string `json:"home_team_name_en"`
	AwayTeamNameEn  string `json:"away_team_name_en"`
	HomeTeamLabel   string `json:"home_team_label,omitempty"`
	AwayTeamLabel   string `json:"away_team_label,omitempty"`
	StadiumID       string `json:"stadium_id"`
}

type GroupStanding struct {
	TeamID string `json:"team_id"`
	MP     string `json:"mp"`
	W      string `json:"w"`
	D      string `json:"d"`
	L      string `json:"l"`
	GF     string `json:"gf"`
	GA     string `json:"ga"`
	GD     string `json:"gd"`
	Pts    string `json:"pts"`
}

type Group struct {
	Name  string          `json:"name"`
	Teams []GroupStanding `json:"teams"`
}

type GroupsResponse struct {
	Groups []Group `json:"groups"`
}

type Team struct {
	ID       string `json:"id"`
	NameEN   string `json:"name_en"`
	FifaCode string `json:"fifa_code"`
	Flag     string `json:"flag"`
	Group    string `json:"groups"`
}

type TeamsResponse struct {
	Teams []Team `json:"teams"`
}
