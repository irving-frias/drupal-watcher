//go:build worldcup

package worldcup

// ─── worldcup26.ir (live 2026) types ───

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

// ─── Openfootball (historical) types ───

type OFTournament struct {
	Name    string    `json:"name"`
	Matches []OFMatch `json:"matches"`
}

type OFMatch struct {
	Round  string   `json:"round"`
	Date   string   `json:"date"`
	Time   string   `json:"time,omitempty"`
	Team1  string   `json:"team1"`
	Team2  string   `json:"team2"`
	Score  *OFScore `json:"score,omitempty"`
	Goals1 []OFGoal `json:"goals1,omitempty"`
	Goals2 []OFGoal `json:"goals2,omitempty"`
	Group  string   `json:"group,omitempty"`
	Ground string   `json:"ground,omitempty"`
	Num    int      `json:"num,omitempty"`
}

type OFScore struct {
	FT []int `json:"ft"`
	HT []int `json:"ht,omitempty"`
	ET []int `json:"et,omitempty"`
	P  []int `json:"p,omitempty"`
}

type OFGoal struct {
	Name    string      `json:"name"`
	Minute  interface{} `json:"minute,omitempty"`
	Penalty bool        `json:"penalty,omitempty"`
	OwnGoal bool        `json:"owngoal,omitempty"`
}

type OFTeamHistoryEntry struct {
	Year          string
	Tournament    string
	Host          bool
	MatchesPlayed int
	Wins          int
	Draws         int
	Losses        int
	GoalsFor      int
	GoalsAgainst  int
	Result        string
}

type OFTeamHistory struct {
	TeamName    string
	Appearances int
	TotalMP     int
	TotalW      int
	TotalD      int
	TotalL      int
	TotalGF     int
	TotalGA     int
	BestResult  string
	Entries     []OFTeamHistoryEntry
}

type OFYear struct {
	Year string
	Name string
}

var KnownWorldCupYears = []OFYear{
	{Year: "1930", Name: "1930 Uruguay"},
	{Year: "1934", Name: "1934 Italy"},
	{Year: "1938", Name: "1938 France"},
	{Year: "1950", Name: "1950 Brazil"},
	{Year: "1954", Name: "1954 Switzerland"},
	{Year: "1958", Name: "1958 Sweden"},
	{Year: "1962", Name: "1962 Chile"},
	{Year: "1966", Name: "1966 England"},
	{Year: "1970", Name: "1970 Mexico"},
	{Year: "1974", Name: "1974 West Germany"},
	{Year: "1978", Name: "1978 Argentina"},
	{Year: "1982", Name: "1982 Spain"},
	{Year: "1986", Name: "1986 Mexico"},
	{Year: "1990", Name: "1990 Italy"},
	{Year: "1994", Name: "1994 USA"},
	{Year: "1998", Name: "1998 France"},
	{Year: "2002", Name: "2002 Korea/Japan"},
	{Year: "2006", Name: "2006 Germany"},
	{Year: "2010", Name: "2010 South Africa"},
	{Year: "2014", Name: "2014 Brazil"},
	{Year: "2018", Name: "2018 Russia"},
	{Year: "2022", Name: "2022 Qatar"},
	{Year: "2026", Name: "2026 Canada/Mexico/USA"},
}
