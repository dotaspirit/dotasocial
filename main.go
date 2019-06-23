package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/hashicorp/go-retryablehttp"
)

type WikiResponse struct {
	Parse struct {
		Text struct {
			ParserOutput string `json:"*"`
		} `json:"text"`
	} `json:"parse"`
}

type Player struct {
	Account_id int    `json:"account_id,omitempty"`
	Facebook   string `json:"facebook,omitempty"`
	Instagram  string `json:"instagram,omitempty"`
	Reddit     string `json:"reddit,omitempty"`
	Steam      string `json:"steam,omitempty"`
	Twitch     string `json:"twitch,omitempty"`
	Twitter    string `json:"twitter,omitempty"`
	Vk         string `json:"vk,omitempty"`
	Weibo      string `json:"weibo,omitempty"`
	YouTube    string `json:"youtube,omitempty"`
}

var players []Player

type ByAccountId []Player

func (a ByAccountId) Len() int           { return len(a) }
func (a ByAccountId) Less(i, j int) bool { return a[i].Account_id < a[j].Account_id }
func (a ByAccountId) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func getJSON(url string, target interface{}) error {
	r, err := retryablehttp.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func getPlayerData(url string) {
	url = "https://liquipedia.net/dota2/api.php?action=parse&prop=text&format=json&page=" + url

	account_id := 0
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
		if link.Find("i", "class", "lp-vkontakte").Pointer != nil {
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

		account_id, _ = strconv.Atoi(dotabuff)
	}

	playerSocialData := Player{
		Account_id: account_id,
		Facebook:   facebook,
		Twitch:     twitch,
		Twitter:    twitter,
		Vk:         vk,
		Weibo:      weibo,
		Instagram:  instagram,
		Reddit:     reddit,
		Steam:      steam,
		YouTube:    youtube,
	}

	if playerSocialData.Account_id != 0 {
		players = append(players, playerSocialData)
	}
}

func getPlayers(url string) {
	url = "https://liquipedia.net/dota2/api.php?action=parse&prop=text&format=json&page=" + url
	var playersData WikiResponse
	getJSON(url, &playersData)
	doc := soup.HTMLParse(playersData.Parse.Text.ParserOutput)
	tables := doc.FindAll("tbody")
	for _, table := range tables {
		rows := table.FindAll("tr")
		for _, row := range rows {
			cols := row.FindAll("td")
			for i, col := range cols {
				if i%5 == 1 {
					playerLink := col.Find("a")
					linkHref := playerLink.Attrs()["href"]
					linkHrefSplit := strings.Split(linkHref, "/")
					lnk := linkHrefSplit[len(linkHrefSplit)-1]
					getPlayerData(lnk)
					time.Sleep(time.Minute)
				}
			}
		}
	}
}

func getPlayersJson() {
	urlList := [4]string{"Portal:Players/Americas", "Portal:Players/Europe", "Portal:Players/China", "Portal:Players/Southeast_Asia"}

	for _, url := range urlList {
		getPlayers(url)
	}

	sort.Sort(ByAccountId(players))

	b, err := json.Marshal(players)
	if err != nil {
		fmt.Println("error:", err)
	}

	f, err := os.Create("players.json")
	f.Write(b)
}

func main() {
	getPlayersJson()
}
