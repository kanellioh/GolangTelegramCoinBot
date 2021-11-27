package main

import (
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type binanceResp struct {
	Price float64 `json:"price,string"`
	Code  int64   `json:"code"`
}

var db = map[int64]wallet{}

type wallet map[string]float64

func main() {
	bot, err := tgbotapi.NewBotAPI("secret")
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		msgArr := strings.Split(update.Message.Text, " ")

		newTextMessage := ""

		switch msgArr[0] {
		case "ADD":
			if len(msgArr) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда"))
				continue
			}
			money, err := strconv.ParseFloat(msgArr[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка"))
				continue
			}

			if _, ok := db[update.Message.Chat.ID]; !ok {
				db[update.Message.Chat.ID] = wallet{}
			}

			db[update.Message.Chat.ID][msgArr[1]] += money

			newTextMessage = fmt.Sprintf("Добавлено на баланс %s:  %f", msgArr[1], db[update.Message.Chat.ID][msgArr[1]])
		case "SUB":
			if len(msgArr) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда"))
				continue
			}
			money, err := strconv.ParseFloat(msgArr[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка"))
				continue
			}

			if _, ok := db[update.Message.Chat.ID]; !ok {
				db[update.Message.Chat.ID] = wallet{}
			}

			db[update.Message.Chat.ID][msgArr[1]] -= money

			newTextMessage = fmt.Sprintf("Добавлено на баланс %s:  %f", msgArr[1], db[update.Message.Chat.ID][msgArr[1]])
		case "DEL":
			was := db[update.Message.Chat.ID][msgArr[1]]
			//delete(db[update.Message.Chat.ID], msgArr[1])
			newTextMessage = fmt.Sprintf("Баланс стал нулевым. На нем было:  %f", was)
		case "SHOW":
			newTextMessage = "Ваш баланс: \n"

			for key, value := range db[update.Message.Chat.ID] {
				dollar, err := getPrice(key, "USD")
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				}
				rub, err := getPrice("USD", "RUB")
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				}

				newTextMessage += fmt.Sprintf("%s: %f ($%.2f) (%.2f рублей)\n", key, value, value*dollar, value*dollar*rub)
			}
		default:
			newTextMessage = "Неизвестная команда"
		}

		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, newTextMessage))
	}
}

func getPrice(coinFrom string, coinTo string) (price float64, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sT%s", coinFrom, coinTo))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var jsonResp binanceResp
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)

	if err != nil {
		return
	}

	if jsonResp.Code != 0 {
		err = errors.New("Ошибка обработки валюты")
		return
	}

	price = jsonResp.Price

	return
}
