package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	Token   	string
	roleID  	string
	ownerID 	string
	link		string
	streamname	string
	allowItPre 	string
	allowIt 	[]string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&roleID, "r", "", "Role-ID")
	flag.StringVar(&ownerID, "o", "", "Owner-ID")
	flag.StringVar(&link, "l", "", "Streamer Link")
	flag.StringVar(&streamname, "s", "", "Streamer Name")
	flag.StringVar(&allowItPre, "a", "", "Allow-It-List")
	flag.Parse()
}

func main() {
	allowIt = strings.Split(allowItPre, ",")
	print("Got allow it")

	for i := range allowIt {
		print(" " + allowIt[i])
	}

	println()

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	dg.AddHandlerOnce(ready)

	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	_ = dg.Close()
}

func ready(s *discordgo.Session, r *discordgo.Ready) {
	embed := NewEmbed().
		SetTitle("Online").
		SetDescription("The bot successfully started up!").
		AddField("Logged in as:", s.State.User.Username).
		AddField("Set RoleID:", roleID).
		AddField("Bot ID:", s.State.User.ID).
		AddField("Allow it List:", allowItPre).
		AddField("Reported Version:", string(rune(r.Version))).
		SetImage(s.State.User.AvatarURL("2048")).
		SetColor(0x00ff00).MessageEmbed
	var sendTo, _ = s.UserChannelCreate(ownerID)
	_, _ = s.ChannelMessageSendEmbed(sendTo.ID, embed)
	_ = s.UpdateListeningStatus("the big chungus song")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !strings.HasPrefix(m.Content, "/") {
		return
	}
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "/twitch" {
		if ComesFromDM(s, m) {
			var sendTo, _ = s.UserChannelCreate(m.Author.ID)
			_, _ = s.ChannelMessageSend(sendTo.ID, "You can't run this in DMs, dummy.")
			return
		}

		if !contains(allowIt, m.Author.ID) {
			var sendTo, _ = s.UserChannelCreate(m.Author.ID)
			_, _ = s.ChannelMessageSend(sendTo.ID, "You think you're funny, huh? Executing the /twitch command.\n" +
				"You're probably thinking \"haha time to mass ping!\" but no, all you did is ridicule yourself. " +
				"Fucking clown.")
			_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
			return
		}

		_, _ = s.GuildRoleEdit(m.GuildID, roleID, "Twitch Notifications", 0xd12de0, true, 0, true)
		time.Sleep(10)
		_, _ = s.ChannelMessageSend(m.ChannelID, "<@&"+roleID+"> " + streamname + " is live! Go watch the stream! " + link)
		_, _ = s.GuildRoleEdit(m.GuildID, roleID, "Twitch Notifications", 0xd12de0, true, 0, false)
	} else if m.Content == "/online" && m.Author.ID == ownerID {
		embed := NewEmbed().
			SetTitle("Online").
			SetDescription("The bot successfully started up!").
			AddField("Logged in as:", s.State.User.Username).
			AddField("Set RoleID:", roleID).
			AddField("Bot ID:", s.State.User.ID).
			AddField("Allow it List:", "NULL").
			SetImage(s.State.User.AvatarURL("2048")).
			SetColor(0x00ff00).MessageEmbed
		_, _ = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ComesFromDM(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		print(err)
		if channel, err = s.Channel(m.ChannelID); err != nil {
			return false
		}
	}

	return channel.Type == discordgo.ChannelTypeDM
}

type Embed struct {
	*discordgo.MessageEmbed
}

const (
	EmbedLimitTitle       = 256
	EmbedLimitDescription = 2048
	EmbedLimitFieldValue  = 1024
	EmbedLimitFieldName   = 256
	EmbedLimitField       = 25
	EmbedLimitFooter      = 2048
)

//NewEmbed returns a new embed object
func NewEmbed() *Embed {
	return &Embed{&discordgo.MessageEmbed{}}
}

//SetTitle ...
func (e *Embed) SetTitle(name string) *Embed {
	e.Title = name
	return e
}

//SetDescription [desc]
func (e *Embed) SetDescription(description string) *Embed {
	if len(description) > 2048 {
		description = description[:2048]
	}
	e.Description = description
	return e
}

//AddField [name] [value]
func (e *Embed) AddField(name, value string) *Embed {
	if len(value) > 1024 {
		value = value[:1024]
	}

	if len(name) > 1024 {
		name = name[:1024]
	}

	e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
		Name:  name,
		Value: value,
	})

	return e

}

//SetFooter [Text] [iconURL]
func (e *Embed) SetFooter(args ...string) *Embed {
	iconURL := ""
	text := ""
	proxyURL := ""

	switch {
	case len(args) > 2:
		proxyURL = args[2]
		fallthrough
	case len(args) > 1:
		iconURL = args[1]
		fallthrough
	case len(args) > 0:
		text = args[0]
	case len(args) == 0:
		return e
	}

	e.Footer = &discordgo.MessageEmbedFooter{
		IconURL:      iconURL,
		Text:         text,
		ProxyIconURL: proxyURL,
	}

	return e
}

//SetImage ...
func (e *Embed) SetImage(args ...string) *Embed {
	var URL string
	var proxyURL string

	if len(args) == 0 {
		return e
	}
	if len(args) > 0 {
		URL = args[0]
	}
	if len(args) > 1 {
		proxyURL = args[1]
	}
	e.Image = &discordgo.MessageEmbedImage{
		URL:      URL,
		ProxyURL: proxyURL,
	}
	return e
}

//SetThumbnail ...
func (e *Embed) SetThumbnail(args ...string) *Embed {
	var URL string
	var proxyURL string

	if len(args) == 0 {
		return e
	}
	if len(args) > 0 {
		URL = args[0]
	}
	if len(args) > 1 {
		proxyURL = args[1]
	}
	e.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL:      URL,
		ProxyURL: proxyURL,
	}
	return e
}

//SetAuthor ...
func (e *Embed) SetAuthor(args ...string) *Embed {
	var (
		name     string
		iconURL  string
		URL      string
		proxyURL string
	)

	if len(args) == 0 {
		return e
	}
	if len(args) > 0 {
		name = args[0]
	}
	if len(args) > 1 {
		iconURL = args[1]
	}
	if len(args) > 2 {
		URL = args[2]
	}
	if len(args) > 3 {
		proxyURL = args[3]
	}

	e.Author = &discordgo.MessageEmbedAuthor{
		Name:         name,
		IconURL:      iconURL,
		URL:          URL,
		ProxyIconURL: proxyURL,
	}

	return e
}

//SetURL ...
func (e *Embed) SetURL(URL string) *Embed {
	e.URL = URL
	return e
}

//SetColor ...
func (e *Embed) SetColor(clr int) *Embed {
	e.Color = clr
	return e
}

// InlineAllFields sets all fields in the embed to be inline
func (e *Embed) InlineAllFields() *Embed {
	for _, v := range e.Fields {
		v.Inline = true
	}
	return e
}

// Truncate truncates any embed value over the character limit.
func (e *Embed) Truncate() *Embed {
	e.TruncateDescription()
	e.TruncateFields()
	e.TruncateFooter()
	e.TruncateTitle()
	return e
}

// TruncateFields truncates fields that are too long
func (e *Embed) TruncateFields() *Embed {
	if len(e.Fields) > 25 {
		e.Fields = e.Fields[:EmbedLimitField]
	}

	for _, v := range e.Fields {

		if len(v.Name) > EmbedLimitFieldName {
			v.Name = v.Name[:EmbedLimitFieldName]
		}

		if len(v.Value) > EmbedLimitFieldValue {
			v.Value = v.Value[:EmbedLimitFieldValue]
		}

	}
	return e
}

// TruncateDescription ...
func (e *Embed) TruncateDescription() *Embed {
	if len(e.Description) > EmbedLimitDescription {
		e.Description = e.Description[:EmbedLimitDescription]
	}
	return e
}

// TruncateTitle ...
func (e *Embed) TruncateTitle() *Embed {
	if len(e.Title) > EmbedLimitTitle {
		e.Title = e.Title[:EmbedLimitTitle]
	}
	return e
}

// TruncateFooter ...
func (e *Embed) TruncateFooter() *Embed {
	if e.Footer != nil && len(e.Footer.Text) > EmbedLimitFooter {
		e.Footer.Text = e.Footer.Text[:EmbedLimitFooter]
	}
	return e
}
