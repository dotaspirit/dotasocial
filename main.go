package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
)

type WikiResponse struct {
	Parse struct {
		Text struct {
			ParserOutput string `json:"*"`
		} `json:"text"`
	} `json:"parse"`
}

type Item struct {
	AccountID int    `json:"account_id,omitempty"`
	TeamID    int    `json:"team_id,omitempty"`
	LeagueID  int    `json:"leagueid,omitempty"`
	Facebook  string `json:"facebook,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	Reddit    string `json:"reddit,omitempty"`
	Steam     string `json:"steam,omitempty"`
	Twitch    string `json:"twitch,omitempty"`
	Twitter   string `json:"twitter,omitempty"`
	Vk        string `json:"vk,omitempty"`
	Weibo     string `json:"weibo,omitempty"`
	YouTube   string `json:"youtube,omitempty"`
}

var items []Item

type ByAccountID []Item

func (a ByAccountID) Len() int           { return len(a) }
func (a ByAccountID) Less(i, j int) bool { return a[i].AccountID < a[j].AccountID }
func (a ByAccountID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByTeamID []Item

func (a ByTeamID) Len() int           { return len(a) }
func (a ByTeamID) Less(i, j int) bool { return a[i].TeamID < a[j].TeamID }
func (a ByTeamID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByLeagueID []Item

func (a ByLeagueID) Len() int           { return len(a) }
func (a ByLeagueID) Less(i, j int) bool { return a[i].LeagueID < a[j].LeagueID }
func (a ByLeagueID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func getJSON(url string, target interface{}) error {
	client := &http.Client{}

	log.Println(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "dotasocial/1.0 (https://yay.qa/; me@yay.qa)")
	req.Header.Set("Accept-Encoding", "gzip")

	r, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer r.Body.Close()

	var reader io.ReadCloser
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(r.Body)
		defer reader.Close()
	default:
		reader = r.Body
	}

	return json.NewDecoder(reader).Decode(target)
}

func getData(url, itemType string) {
	url = "https://liquipedia.net/dota2/api.php?action=parse&prop=text&format=json&page=" + url

	accountID := 0
	teamID := 0
	leagueID := 0
	dotabuff := ""
	twitter := ""
	facebook := ""
	twitch := ""
	vk := ""
	weibo := ""
	instagram := ""
	reddit := ""
	steam := ""
	youtube := ""

	var playerData WikiResponse
	getJSON(url, &playerData)
	doc := soup.HTMLParse(playerData.Parse.Text.ParserOutput)

	infoBox := doc.Find("div", "class", "infobox-center")

	if infoBox.Pointer != nil {
		links := infoBox.FindAll("a")
		for _, link := range links {
			linkHref := link.Attrs()["href"]
			linkHrefSplit := strings.Split(linkHref, "/")
			lnk := linkHrefSplit[len(linkHrefSplit)-1]
			if link.Find("i", "class", "lp-dotabuff").Pointer != nil {
				dotabuff = lnk
			}
			if link.Find("i", "class", "lp-twitter").Pointer != nil {
				twitter = lnk
			}
			if link.Find("i", "class", "lp-facebook").Pointer != nil {
				facebook = lnk
			}
			if link.Find("i", "class", "lp-twitch").Pointer != nil {
				twitch = lnk
			}
			if link.Find("i", "class", "lp-vk").Pointer != nil {
				vk = lnk
			}
			if link.Find("i", "class", "lp-weibo").Pointer != nil {
				weibo = lnk
			}
			if link.Find("i", "class", "lp-instagram").Pointer != nil {
				instagram = lnk
			}
			if link.Find("i", "class", "lp-reddit").Pointer != nil {
				reddit = lnk
			}
			if link.Find("i", "class", "lp-steam").Pointer != nil {
				steam = lnk
			}
			if link.Find("i", "class", "lp-youtube").Pointer != nil {
				youtube = lnk
			}

			if itemType == "player" {
				accountID, _ = strconv.Atoi(dotabuff)
			} else if itemType == "team" {
				teamID, _ = strconv.Atoi(dotabuff)
			} else if itemType == "league" {
				leagueID, _ = strconv.Atoi(dotabuff)
			}
		}
	}

	itemSocialData := Item{
		AccountID: accountID,
		TeamID:    teamID,
		LeagueID:  leagueID,
		Facebook:  facebook,
		Twitch:    twitch,
		Twitter:   twitter,
		Vk:        vk,
		Weibo:     weibo,
		Instagram: instagram,
		Reddit:    reddit,
		Steam:     steam,
		YouTube:   youtube,
	}

	if itemSocialData.AccountID != 0 ||
		itemSocialData.LeagueID != 0 ||
		itemSocialData.TeamID != 0 {
		items = append(items, itemSocialData)
	}
}

func getPlayers(url string) {
	url = "https://liquipedia.net/dota2/api.php?action=parse&prop=text&format=json&page=" + url
	var playersData WikiResponse
	getJSON(url, &playersData)
	doc := soup.HTMLParse(playersData.Parse.Text.ParserOutput)
	blockPlayer := doc.FindAll("div", "class", "block-player")
	fmt.Println(blockPlayer)
	for _, player := range blockPlayer {
		playerLink := player.Find("a")
		linkHref := playerLink.Attrs()["href"]
		linkHrefSplit := strings.Split(linkHref, "/")
		lnk := linkHrefSplit[len(linkHrefSplit)-1]
		getData(lnk, "player")
		time.Sleep(time.Minute)
	}
}

func getTeams(url string) {
	url = "https://liquipedia.net/dota2/api.php?action=parse&prop=text&format=json&page=" + url
	var teamsData WikiResponse
	getJSON(url, &teamsData)
	doc := soup.HTMLParse(teamsData.Parse.Text.ParserOutput)
	teams := doc.FindAll("span", "class", "team-template-text")
	for _, team := range teams {
		teamLink := team.Find("a")
		linkHref := teamLink.Attrs()["href"]
		linkHrefSplit := strings.Split(linkHref, "/")
		lnk := linkHrefSplit[len(linkHrefSplit)-1]
		getData(lnk, "team")
		time.Sleep(time.Minute)
	}
}

func getLeagues(url string) {
	url = "https://liquipedia.net/dota2/api.php?action=parse&prop=text&format=json&page=" + url
	var leaguesData WikiResponse
	getJSON(url, &leaguesData)
	doc := soup.HTMLParse(leaguesData.Parse.Text.ParserOutput)
	leagues := doc.FindAllStrict("div", "class", "gridCell Tournament Header")
	for _, league := range leagues {
		leagueLink := league.FindAll("a")
		linkHref := leagueLink[len(leagueLink)-1].Attrs()["href"]
		lnk := strings.Replace(linkHref, "/dota2/", "", -1)
		getData(lnk, "league")
		time.Sleep(time.Minute)
	}
}

func getPlayersJson() {
	urlList := [4]string{"Portal:Players/Americas", "Portal:Players/Europe", "Portal:Players/China", "Portal:Players/Southeast_Asia"}

	for _, url := range urlList {
		getPlayers(url)
		time.Sleep(time.Minute)
	}

	sort.Sort(ByAccountID(items))

	b, err := json.Marshal(items)
	if err != nil {
		fmt.Println("error:", err)
	}

	f, err := os.Create("players.json")
	f.Write(b)
}

func getTeamsJson() {
	urlList := [4]string{"Portal:Teams"}

	for _, url := range urlList {
		getTeams(url)
		time.Sleep(time.Minute)
	}

	sort.Sort(ByTeamID(items))

	b, err := json.Marshal(items)
	if err != nil {
		fmt.Println("error:", err)
	}

	f, err := os.Create("teams.json")
	f.Write(b)
}

func getLeaguesJson() {
	urlList := [4]string{"Tier_1_Tournaments", "Tier_2_Tournaments/2022-2023"}

	for _, url := range urlList {
		getLeagues(url)
		time.Sleep(time.Minute)
	}

	sort.Sort(ByLeagueID(items))

	b, err := json.Marshal(items)
	if err != nil {
		fmt.Println("error:", err)
	}

	f, err := os.Create("leagues.json")
	f.Write(b)
}

func main() {
	items = nil
	getTeamsJson()
	items = nil
	getLeaguesJson()
	items = nil
	getPlayersJson()
}
