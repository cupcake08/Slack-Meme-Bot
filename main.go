package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/joho/godotenv"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
)

const MEME_API = "https://meme-api.com/gimme"

type Meme struct {
	Link string `json:"postLink"`
	Url string `json:"url"`
	Author string `json:"author"`
	Title string `json:"title"`
}

func printCommandEvents(analytics <- chan *slacker.CommandEvent) {
	for event := range analytics {
		fmt.Println("Command Events")
		fmt.Println(event.Timestamp)
		fmt.Println(event.Command)
		fmt.Println(event.Parameters)
		fmt.Println(event.Event)
		fmt.Println()
	}
}

func getImage() (*Meme,string,error) {
	res, err := http.Get(MEME_API)
	if err != nil {
		return nil,"",err 
	}
	defer res.Body.Close()
	body,err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil,"",err 
	}
	meme := &Meme{}
	json.Unmarshal(body,meme)

	res, err = http.Get(meme.Url)
	if err != nil {
		return nil,"",err 
	}
	defer res.Body.Close()

	fname := path.Base(meme.Url)
	f,err := os.Create(fname)

	defer f.Close()

	_,err = f.ReadFrom(res.Body)
	if err != nil {
		return nil,"",err 
	}

	fmt.Println("MEME downloaded.. <:)>")
	return meme,fname,nil 
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err.Error())
	}

	bot := slacker.NewClient(os.Getenv("SLACK_BOT_TOKEN"),os.Getenv("SLACK_APP_TOKEN"))
	client := slack.New(os.Getenv("SLACK_BOT_TOKEN"))
	channels := []string{os.Getenv("CHANNEL_ID")}

	go printCommandEvents(bot.CommandEvents())

	bot.Command("meme",&slacker.CommandDefinition{
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			meme,fname,err := getImage()
			if err != nil {
				log.Fatal(err.Error())
				response.Reply("Something Went Wrong!")
			}else {
				s := fmt.Sprintf("Author: %s\n%s",meme.Author,meme.Title)
				response.Reply(s)
				// upload file
				f,err := client.UploadFile(slack.FileUploadParameters{
					File: fname,
					Channels: channels,
				})
				if err != nil {
					log.Fatal(err.Error())
				}
				fmt.Println("File Uploaded:",f.Name)
				os.Remove(fname)
			}
		},
	})

	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()

	err = bot.Listen(ctx)

	if err != nil {
		log.Fatal(err.Error())
	}
}
