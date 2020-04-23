package main

import (
	"encoding/json"
	"fmt"
	"github.com/rusgreen/whdisco/wh"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const discordChannelId = "ID"                                                                                      // ID канала Discord Webhook для отправки информационных уведомлений
const discordToken = "токен"                                                                                       //	токен Discord Webhook для отправки информационных уведомлений
const discordWhUrl = "https://discordapp.com/api/webhooks/" + discordChannelId + "/" + discordToken                // url Discord Webhook для отправки информационных уведомлений
const errorDiscordChannelId = "ID"                                                                                 // ID канала Discord Webhook для отправки уведомлений об ошибке
const errorDiscordToken = "токен"                                                                                  //	токен Discord Webhook для отправки уведомлений об ошибке
const errorDiscordWhUrl = "https://discordapp.com/api/webhooks/" + errorDiscordChannelId + "/" + errorDiscordToken // url Discord Webhook для отправки уведомлений об ошибке
const urlCountries = "https://corona.lmao.ninja/v2/countries?sort=cases"                                           //	url источника
const pauseBetweenMesages = 1                                                                                      // пауза между сообщениями

type Info struct {
	Country     string `json:"country"`
	Cases       int    `json:"cases"`
	TodayCases  int    `json:"todayCases"`
	Deaths      int    `json:"deaths"`
	TodayDeaths int    `json:"todayDeaths"`
	Recovered   int    `json:"recovered"`
	Active      int    `json:"active"`
	Critical    int    `json:"critical"`
	Number      int
}

var err error
var responseCountries *http.Response
var informCountries []uint8

func main() {
	var previousCountries []*Info
	for {
		for {
			responseCountries, err = http.Get(urlCountries)
			if err != nil {
				sendErrorWebhooks(err)
			} else {
				informCountries, err = ioutil.ReadAll(responseCountries.Body)
				if err != nil {
					sendErrorWebhooks(err)
				} else if responseCountries != nil && informCountries != nil {
					break
				}
			}
			time.Sleep(30 * time.Second)
		}
		// десериализация информации по всем старанам мира
		var countriesData []*Info
		err = json.Unmarshal(informCountries, &countriesData)
		if err != nil {
			sendErrorWebhooks(err)
		} else {
			currentCountries := countriesData
			// добавляем нумерацию
			for i, v := range currentCountries {
				v.Number = i + 1
			}
			// определяем разницу между текущим и предыдущим запросом
			diffResult := difference(currentCountries, previousCountries)
			// собираем необходимую информацию и отправляем её посредством Discord Webhook
			BuildAndSendWebhooks(diffResult, previousCountries)
			previousCountries = currentCountries
		}
		time.Sleep(10 * time.Minute)
	}
}

func difference(slice1 []*Info, slice2 []*Info) []*Info {
	var diff []*Info
	for _, s1 := range slice1 {
		changed := false
		for _, s2 := range slice2 {
			if s1.Country == s2.Country && s1.Cases != s2.Cases {
				changed = true
				break
			}
		}
		if changed {
			diff = append(diff, s1)
		}
	}
	return diff
}

func BuildAndSendWebhooks(diffResult []*Info, previousCountries []*Info) {
	webhook := wh.NewDiscordWebhook(discordWhUrl)
	webhook.SetStatusYellow()
	if len(diffResult) > 0 {
		sliceOfDescriptions := make([]string, 0)
		for _, diffSlice := range diffResult {
			for _, previousSlice := range previousCountries {
				if diffSlice.Country == previousSlice.Country {
					changeCases := diffSlice.Cases - previousSlice.Cases
					changeDeaths := diffSlice.Deaths - previousSlice.Deaths
					changeRecovered := diffSlice.Recovered - previousSlice.Recovered
					var deltaCases string
					var deltaDeaths string
					var deltaRecovered string
					if changeCases > 0 {
						deltaCases = fmt.Sprintf(" (+%v)", changeCases)
					} else if changeCases < 0 {
						deltaCases = fmt.Sprintf(" (%v)", changeCases)
					}
					if changeDeaths > 0 {
						deltaDeaths = fmt.Sprintf(" (+%v)", changeDeaths)
					} else if changeDeaths < 0 {
						deltaDeaths = fmt.Sprintf(" (%v)", changeDeaths)
					}
					if changeRecovered > 0 {
						deltaRecovered = fmt.Sprintf(" (+%v)", changeRecovered)
					} else if changeRecovered < 0 {
						deltaRecovered = fmt.Sprintf(" (%v)", changeRecovered)
					}
					description := fmt.Sprintf("**%s** №%v\n Заражений всего: %v"+deltaCases+"\n Заражённых за сутки: %v\n Погибших всего: %v"+deltaDeaths+"\n Погибших за сутки: %v\n Заражённых сейчас: %v\n В критическом состоянии: %v\n Выздоровевших: %v"+deltaRecovered+"\n\n", diffSlice.Country, diffSlice.Number, diffSlice.Cases, diffSlice.TodayCases, diffSlice.Deaths, diffSlice.TodayDeaths, diffSlice.Active, diffSlice.Critical, diffSlice.Recovered)
					sliceOfDescriptions = append(sliceOfDescriptions, description)
				}
			}
		}
		switch {
		case len(diffResult) <= 10:
			firstMessage := strings.Join(sliceOfDescriptions, "")
			webhook.SetDescription(fmt.Sprintln(firstMessage))
			err = webhook.Send()
			if err != nil {
				fmt.Println(err)
			}
		case len(diffResult) > 10 && len(diffResult) <= 20:
			firstMessage := strings.Join(sliceOfDescriptions[:10], "")
			secondMessage := strings.Join(sliceOfDescriptions[10:20], "")
			webhook.SetDescription(fmt.Sprintln(firstMessage))
			err = webhook.Send()
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(pauseBetweenMesages * time.Second)
			webhook.SetDescription(fmt.Sprintln(secondMessage))
			err = webhook.Send()
			if err != nil {
				fmt.Println(err)
			}
		case len(diffResult) > 20 && len(diffResult) <= 30:
			firstMessage := strings.Join(sliceOfDescriptions[:10], "")
			secondMessage := strings.Join(sliceOfDescriptions[10:20], "")
			thirdMessage := strings.Join(sliceOfDescriptions[20:30], "")
			webhook.SetDescription(fmt.Sprintln(firstMessage))
			err = webhook.Send()
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(pauseBetweenMesages * time.Second)
			webhook.SetDescription(fmt.Sprintln(secondMessage))
			err = webhook.Send()
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(pauseBetweenMesages * time.Second)
			webhook.SetDescription(fmt.Sprintln(thirdMessage))
			err = webhook.Send()
			if err != nil {
				fmt.Println(err)
			}
		case len(diffResult) > 30 && len(diffResult) <= 40:
			firstMessage := strings.Join(sliceOfDescriptions[:10], "")
			secondMessage := strings.Join(sliceOfDescriptions[10:20], "")
			thirdMessage := strings.Join(sliceOfDescriptions[20:30], "")
			fourthMessage := strings.Join(sliceOfDescriptions[30:40], "")
			webhook.SetDescription(fmt.Sprintln(firstMessage))
			err = webhook.Send()
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(pauseBetweenMesages * time.Second)
			webhook.SetDescription(fmt.Sprintln(secondMessage))
			err = webhook.Send()
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(pauseBetweenMesages * time.Second)
			webhook.SetDescription(fmt.Sprintln(thirdMessage))
			err = webhook.Send()
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(pauseBetweenMesages * time.Second)
			webhook.SetDescription(fmt.Sprintln(fourthMessage))
			err = webhook.Send()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func sendErrorWebhooks(err error) {
	webhook := wh.NewDiscordWebhook(errorDiscordWhUrl)
	webhook.SetStatusRed()
	webhook.SetDescription("Источник: " + urlCountries + "\nОшибка: " + err.Error())
	err = webhook.Send()
	if err != nil {
		fmt.Println(err)
	}
}
