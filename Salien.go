/*
------------------------------------------------------------
2018 Summer Saliens - Bosswatch
                                      I'm hunting wabbits...
------------------------------------------------------------

This app polls Steam's servers requesting for json-replies about the state of
the game and focuses on finding those tasty bosses out of the currently active
planets. When the app finds active boss encounter, it will print alert for user
where to find this boss and after the boss dies, it aproximates the encounter
duration and estimate of maximum amount of EXP user can gain in that time based
on the encounter length.

Totally useless after the even goes away (July 4th at 10am PST).

If you are after some fast copy&paste -code for your school/course work, I
suggest you copy from someone with better knowhow than me for your own sake,
because I'm not an expert, just hobbyist and this is my first Go-project.
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	POLLING_INTERVAL = 15 // Don't hammer the Steam API too much, 10 is probably safe, and I wouldn't go under 6, this app isn't that time-relied.
	EXP_PER_TICK = 2500 // You get around 2500? EXP for every 5 second tick (12 ticks in minute) you are in boss encounter
	SUPPORTED_GAMEVERSION = 2 // Just incase Valve updates something, we get notification through crash and burn
	TIMESTAMP = "02.01.2006 15:04:05" // Timestamps, so much fun in Go

	// https://community.steam-api.com/ITerritoryControlMinigameService/GetPlanets/v0001/?language=english
	// https://community.steam-api.com/ITerritoryControlMinigameService/GetPlanet/v0001/?&language=english&id=[[planet.id]]
	BASE_URL = "https://community.steam-api.com/ITerritoryControlMinigameService/"
	END_URL = "/v0001/?language=english"
)

type State struct {
	Name				string		`json:"name"`
	Active				bool		`json:"active"`
	Captured			bool		`json:"captured"`
	CaptureProgress		float64		`json:"capture_progress"` // We don't really need this
	BossZonePosition	int			`json:"boss_zone_position"`
}

type Zones struct {
	ZonePosition		int			`json:"zone_position"` // Not needed since we use BossZonePosition now
	Type				int			`json:"type"` // Not needed since we use BossZonePosition now
	BossActive			bool		`json:"boss_active"`
}

type Planets struct {
	Id					string		`json:"id"`
	State				State		`json:"state"`
	Zones				[]Zones		`json:"zones"`
}

type Response struct {
	GameVersion			int			`json:"game_version"`
	Planets				[]Planets	`json:"planets"`
}

type SalienAPIResponse struct {
	Response			Response	`json:"response"`
}

func lpad(s string, pad string, plength int) string { // Not written by me, found it from Interwebs and added few spaces here and there
	for i := len(s); i < plength; i++ {
		s = pad + s
	}
	return s
}

func getUrl(s string) string { // Wrote this just to keep the line length shorter on NewRequest-lines
	return BASE_URL + s + END_URL
}

// Setting up variables
var startScript, startBoss, startIteration, now, lastBossKilled time.Time
var bossFightActive, bossFightRegistered, errored bool
var bossCount, planetCount int

func main() {
	salienClient := http.Client{
		//Timeout: time.Second * 2, // Maximum of 2 secs
		Timeout: time.Second * POLLING_INTERVAL / 2, // Steam's servers are on high load from time to time...
	}
	startScript = time.Now()
	bossFightRegistered = false // Set initial value to false
	errored = false // Set initial value to false
	bossCount = 0 // Set initial value to 0

	// Small banner style start up and timestamp
	fmt.Printf("%v\n2018 Summer Saliens - Bosswatch\n%v\n%v\n\n", lpad("-", "-", 60), lpad("I'm hunting wabbits...", " ", 60), lpad("-", "-", 60))
	fmt.Printf("[%v] Script started.\n", startScript.Format(TIMESTAMP))

	TryToSelfFix: // In case we get timeout error, don't die but wait and iterate again
	if errored == true {
		fmt.Println("\n" + lpad("!", "!", 60))
		fmt.Printf("[%v] >>> Error, waiting little bit and trying to continue...\n", time.Now().Format(TIMESTAMP))
		time.Sleep(time.Second * POLLING_INTERVAL * 2)
		fmt.Printf("[%v] >>> Continuing...\n", time.Now().Format(TIMESTAMP))
		errored = false
	}

	for {
		// while true -loop

		startIteration = time.Now()
		bossFightActive = false // Set initial value to false
		planetCount = 0 // Count planets to detect if we captured them all

		// Create http-request for getting list of planets
		req, err := http.NewRequest(http.MethodGet, getUrl("GetPlanets"), nil)
		if err != nil {
			fmt.Println("\n! P-REQ: " + err.Error())
			errored = true
			goto TryToSelfFix
			//log.Fatal(err)
		}

		// Alter Headers so we look like Browser. Valve doesn't care so this is just to protect user from ISP sniffing Headers and extra sugar on top.
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.87 Safari/537.36")
		req.Header.Set("Referer", "https://steamcommunity.com/saliengame/play/")
		req.Header.Set("Origin", "https://steamcommunity.com")

		// Send http-request
		res, getErr := salienClient.Do(req)
		if getErr != nil {
			fmt.Println("\n! P-RES: " + getErr.Error())
			errored = true
			goto TryToSelfFix
			//log.Fatal(getErr)
		}

		// Read results
		body, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			log.Fatal("\n! P-READ: " + readErr.Error())
		}

		// Parse json
		p:= new(SalienAPIResponse)
		jsonErr := json.Unmarshal(body, &p)
		if jsonErr != nil {
			log.Fatal("\n! P-JSON: " + jsonErr.Error())
		}

		// Validate GameVersion to make sure we know what to do with it.
		if p.Response.GameVersion != SUPPORTED_GAMEVERSION {
			log.Fatal("GameVersion missmatch on PLANETS: Expected ", SUPPORTED_GAMEVERSION, ", got ", p.Response.GameVersion, "\n")
		}

		// Iterate through all active planets (non-active and already captured planets are skipped).
		for _, Planet := range p.Response.Planets {
			if Planet.State.Active == true && Planet.State.Captured == false {
				planetCount += 1 // Active planet found

				// Create http-request for getting zones of invidual planet
				req, err := http.NewRequest(http.MethodGet, getUrl("GetPlanet") + "&id=" + Planet.Id, nil)
				if err != nil {
					fmt.Println("\n! Z-REQ: " + err.Error())
					errored = true
					goto TryToSelfFix
					//log.Fatal(err)
				}

				// Once again, sprinkling some sugar...
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.87 Safari/537.36")
				req.Header.Set("Referer", "https://steamcommunity.com/saliengame/play/")
				req.Header.Set("Origin", "https://steamcommunity.com")

				// Send http-request
				res, getErr := salienClient.Do(req)
				if getErr != nil {
					fmt.Println("\n! Z-RES: " + getErr.Error())
					errored = true
					goto TryToSelfFix
					//log.Fatal(getErr)
				}

				// Read results
				body, readErr := ioutil.ReadAll(res.Body)
				if readErr != nil {
					log.Fatal("\n! Z-READ: " + readErr.Error())
				}

				// Parse json
				z := new(SalienAPIResponse)
				jsonErr := json.Unmarshal(body, &z)
				if jsonErr != nil {
					log.Fatal("\n! Z-JSON: " + jsonErr.Error())
				}

				// Validate GameVersion once again eventhough we got this far with previous request.
				if z.Response.GameVersion != SUPPORTED_GAMEVERSION {
					log.Fatal("GameVersion missmatch on ZONES: Expected ", SUPPORTED_GAMEVERSION, ", got ", z.Response.GameVersion, "\n")
				}

				// No need to iterate zones if z.Response.Planets[0].State gives me all the information I need?
				/*for _, Zone := range z.Response.Planets[0].Zones {
					if Zone.BossActive == true && Zone.Type == 4 {
						bossFightActive = true
						if bossFightRegistered == false {
							startBoss = time.Now()
							bossFightRegistered = true
							fmt.Printf("Fight started @ %v on planet %v in zone %v!\n", time.Now(), Planet.State.Name, Zone.ZonePosition)
						}
						fmt.Printf(">>>   BOSS ACTIVE AT %v IN ZONE %v   <<<\n", Planet.State.Name, Zone.ZonePosition)
					}
				}*/

				// Cleaner and shorten lines and mostly faster to type
				Boss := z.Response.Planets[0]

				// Check if the planet has BossZonePosition set and if the Boss is active in that Zone just to be sure we don't cry a wolf here...
				// Go won't let me split this line into multiline if-test, so it has to be this one line with 177 characters + indentation...
				if ((Boss.State.BossZonePosition != 0 && Boss.Zones[Boss.State.BossZonePosition].BossActive == true) || (Boss.State.BossZonePosition == 0 && Boss.Zones[0].BossActive == true)) {
					bossFightActive = true
					// We detected new active Boss -> Alert user and set start point for encounter duration counter.
					if bossFightRegistered == false {
						startBoss = time.Now()
						bossFightRegistered = true
						bossCount += 1
						fmt.Println("\n" + lpad("-", "-", 60))
						fmt.Printf("[%v] >>> BOSS DETECTED <<<\n\n- Zone %v on planet '%v' (Id: %v)\n", startBoss.Format(TIMESTAMP), Boss.State.BossZonePosition, Planet.State.Name, Planet.Id)
						if bossCount > 1 {
							fmt.Printf("- %v bosses detected in %v\n", bossCount, startBoss.Sub(startScript).Round(time.Second))
							fmt.Printf("- Last boss killed %v ago\n", startBoss.Sub(lastBossKilled).Round(time.Second))
							fmt.Printf("- Average wait time between bosses is %v", (startBoss.Sub(startScript) / time.Duration(bossCount)).Round(time.Second))
						} else {
							fmt.Printf("- First boss detected %v after the start\n", startBoss.Sub(startScript).Round(time.Second))
						}
					}
				}
			}
		}

		now = time.Now()

		// Check if we have encouter duration counter going on, but the boss has disappeared.
		if bossFightRegistered == true && bossFightActive == false {
			// Reset back to normal state and inform user about the encounter duration aproximation.
			bossFightRegistered = false
			lastBossKilled = now
			encounterLenght := now.Sub(startBoss).Round(time.Second)
			experienceEst := EXP_PER_TICK * encounterLenght.Seconds() / 5
			fmt.Println("\n" + lpad("-", "-", 60))
			fmt.Printf("[%v] <<< BOSS IS GONE >>>\n\n- Encounter lasted ~%v\n- Estimate for maximum EXP gain is %v + bonuses\n", now.Format(TIMESTAMP), encounterLenght, experienceEst)
		}
		if planetCount == 0 { // No active planets found, we must have conquered them all! Ready to exit then
			fmt.Printf("[%v] === ALL PLANETS HAVE BEEN CONQUERED ===\n- %v bosses detected in %v\n- Average wait time between bosses was %v", now.Format(TIMESTAMP), bossCount, startBoss.Sub(startScript).Round(time.Second), (startBoss.Sub(startScript) / time.Duration(bossCount)).Round(time.Second))
			goto ExitPoint
		}

		// Sleep until we have filled the POLLING_INTERVAL
		elapsed := now.Sub(startIteration)
		time.Sleep(POLLING_INTERVAL * time.Second - elapsed)

		// End of while true -loop
	}

	ExitPoint:
}