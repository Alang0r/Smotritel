package main

import (
	"Bot/lib"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-ini/ini"
)

var (
	exit       bool
	lastUpd    = 0
	log        lib.Loger
	movieCount = 0
	context    = ""
	tgToken    string
	adminID int
)

//init настраивает конфигурацию для бота
func init() {
	log.Init()
	cfg, err := ini.Load(lib.IniPath)
	if err != nil {
		log.Println("Fail to read file:", err)
		os.Exit(1)
	}
	tgToken = cfg.Section("telegram").Key("token").String()
	adminID, err = cfg.Section("telegram").Key("adminID").Int()
	if err != nil {
		log.Println(err)
	}
	lastUpd, err = cfg.Section("telegram").Key("last_update_id").Int()
	if err != nil {
		log.Println(err)
	}

}

func main() {
	for {
		//4.1. Проверям сообщения, если есть - обрабатываем
		CheckUpdates()
		time.Sleep(1 * time.Second)
		if exit {
			break
		}
	}
}

//GetMe ...
func GetMe() {
	body := getBodyByURL(getUrlbyMethod(lib.MGetMe), []byte(""))
	getMe := lib.GetMeT{}
	err := json.Unmarshal(body, &getMe)
	if err != nil {
		log.Println(err.Error())
	}
	fmt.Println(getMe.Result)
}

//SendMessage ..;
func SendMessage(id int, text string) {
	message := lib.SMessageT{}
	message.ChatID = id
	message.Text = text
	var jsonStr, err = json.Marshal(message)
	body := getBodyByURL(getUrlbyMethod(lib.MSendMessage), jsonStr)
	resp := lib.MessageT{}
	err1 := json.Unmarshal(body, &resp)
	if err != nil {
		log.Println(err1.Error())
	}

}
func getBodyByURL(url string, data []byte) []byte {
	r := bytes.NewReader(data)
	response, err := http.Post(url, "application/json", r)
	if err != nil {
		log.Println(err.Error())
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err.Error())
	}
	return body

}
func getUrlbyMethod(mName string) string {
	var link = lib.TgBaseURL + tgToken + "/" + mName
	//fmt.Println(link)
	return link
}

//CheckUpdates цикл в котором проверяем новые сообщения, если видим сообщение, обрабатываем
func CheckUpdates() {
	reqParametres := lib.GetSomeUpdatesT{}
	reqParametres.Offset = lastUpd + 1
	var jsonParam, err = json.Marshal(reqParametres)
	body := getBodyByURL(getUrlbyMethod(lib.MGetUpdates), jsonParam)
	getUpd := lib.GetUpdatesT{}
	err1 := json.Unmarshal(body, &getUpd)
	if err1 != nil {
		log.Println(err.Error())
	}
	for _, item := range getUpd.Result {
		lastUpd = item.UpdateID
		SaveState(item.UpdateID, movieCount)
		logStr := "@" + item.Message.From.UserName + ": " + item.Message.Text
		log.Println(logStr)
		parseMessage(item)
	}
}

/*func parseMessage парсим сообщение, чтобы понять что делать.
Примеры: /random /comedy /western /actor Брюс Уиллис /year 2000 /random 5
Выводить ссылку на кинопоиск, либо формировать картинку с обложкой с кинопоиска,
кратким описанием, рейтингом, списком актеров и т.д.*/
func parseMessage(upd lib.GetUpdatesResultT) {
	switch message := upd.Message.Text; message {
	case "/start":
		{
			text := upd.Message.From.FirstName + ", добро пожаловать! Этот бот поможет тебе найти фильм " +
				"для просмотра в ближайший вечер, или сохранить на будущее.\nСписок доступных комманд:" +
				"\n/start - описание бота.\n/random - получить случайный фильм." +
				"\n/random5 - получить 5 случайных фильмов.\n/last5 - получить 5 последних добавленных фильмов." +
				"\nАвтор: @Alexander_G0"
			SendMessage(upd.Message.From.ID, text)
		}
	case "/stop":
		{
			if upd.Message.From.ID == adminID {
				exit = true
			} else {
				SendMessage(upd.Message.From.ID, "Действие доступно только администратору бота:(")
			}

		}
	case "/random":
		{
			//film := "Случайный фильм из моей говноподборки: " + gdocsapi.ReadData()
			m := lib.Movie{}
			m.GetRandom()
			sFilm := MovieToString(m)
			mes := "Случайный фильм из моей говноподборки: " + sFilm
			SendMessage(upd.Message.From.ID, mes)
		}
	case "/random5":
		{
			for i := 0; i < 5; i++ {
				m := lib.Movie{}
				m.GetRandom()
				sFilm := MovieToString(m)
				mes := "Случайный фильм из моей говноподборки: " + sFilm
				SendMessage(upd.Message.From.ID, mes)
			}
			log.Println("отправляем 5 случайных", upd.Message.From.FirstName)
		}
	case "/last5":
		{
			log.Println("отправляем 5 последних ", upd.Message.From.FirstName)
			SendMessage(upd.Message.From.ID, "Последние 5 добавленных:")

			for i := 0; i < 5; i++ {
				var m lib.Movie
				m.GetById(m.Count() - i)
				ms := MovieToString(m)
				SendMessage(upd.Message.From.ID, ms)

			}
		}
	case "/add":
		{
			context = "adding"
			if upd.Message.From.ID == adminID {
				SendMessage(upd.Message.From.ID, "Для добавления фильма отправьте мне его данные в формате:"+
					"\nНазвание_фильма/год/жанр/актёры/рейтинг/комментарий"+
					"\nВ случае отсутствия одного из параметров, оставьте его пустым, сохраняя структуру сообщения"+
					"\nПример: Игра престолов/2012//Лена Хиди")

			}
		}
	default:
		{
			switch context {
			case "adding":
				{
					movie := lib.Movie{}
					parts := strings.Split(message, "/")
					for i, v := range parts {
						if i == 0 {
							movie.Title = v
						}
						if i == 1 {
							if v != "" {
								movie.Year, _ = strconv.Atoi(v)
							}
						}
						if i == 2 {
							movie.Genre = v
						}
						if i == 3 {
							movie.Actors = v
						}
						if i == 4 {
							if v != "" {
								movie.Rating, _ = strconv.Atoi(v)
							}
						}
						if i == 5 {
							movie.Comment = v
						}
					}
					movie.Add()
					ms := "добавлен фильм"+MovieToString(movie)
					SendMessage(upd.Message.From.ID, ms)
					context = ""

				}
			case "":
				{

				}
			default:
				{
					SendMessage(upd.Message.From.ID, "Я Вас не понимаю, ознакомьтесь, пожалуйста, с инструкцией по запросу /start.")
				}
			}
		}

	}

}

//SaveState syncing last updates
func SaveState(_lastUpdateID, _lastRow int) {
	cfg, err := ini.Load(lib.IniPath)
	if err != nil {
		log.Println("Fail to read file: ", err)
		os.Exit(1)
	}
	cfg.Section("telegram").Key("last_update_id").SetValue(strconv.Itoa(_lastUpdateID))
	cfg.SaveTo(lib.IniPath)
}

//MovieToString переделывает структуру фильма в строку для отправки пользователю
func MovieToString(movie lib.Movie) string {
	str := "\nНазвание: "
	str += movie.Title
	if movie.Year != 0 {
		str += "\nГод: " + strconv.Itoa(movie.Year)
	}
	if movie.Actors != "" {
		str += "\nВ ролях: " + movie.Actors
	}
	if movie.Genre != "" {
		str += "\nЖанр: " + movie.Genre
	}
	if movie.Rating != 0 {
		str += "\nОценка: " + strconv.Itoa(movie.Rating)
	}
	if movie.Comment != "" {
		str += "\nОтзыв: " + movie.Comment
	}
	return str
}
