package telegram

import (
	"LinkBot/lib/e"
	"LinkBot/storage"
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
)

var (
	RndCmd   = "/rnd"
	AllCmd   = "/all"
	ListCmd  = "/list"
	HelpCmd  = "/help"
	StartCmd = "/start"
)

const telegramMessageLimit = 4000

func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new command '%s' from '%s'", text, username)

	if isAddCmd(text) {
		return p.savePage(chatID, text, username)
	}

	switch text {
	case RndCmd:
		return p.sendRandom(chatID, username)
	case AllCmd, ListCmd:
		return p.sendAll(chatID, username)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendStart(chatID)
	default:
		return p.tg.SendMessage(chatID, msgUnknownCommand)
	}

}

func (p *Processor) savePage(chatID int, pageURL string, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: save page", err) }()

	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
	}

	isExists, err := p.storage.Exists(context.Background(), page)
	if err != nil {
		return err
	}

	if isExists {
		return p.tg.SendMessage(chatID, msgAlreadyExists)
	}

	if err := p.storage.Save(context.Background(), page); err != nil {
		return err
	}

	if err := p.tg.SendMessage(chatID, msgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendRandom(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't pick random page", err) }()

	page, err := p.storage.PickRandom(context.Background(), username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}
	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatID, msgNoSavedPages)
	}

	if err := p.tg.SendMessage(chatID, page.URL); err != nil {
		return err
	}

	return p.storage.Remove(context.Background(), page)
}

func (p *Processor) sendAll(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't list pages", err) }()

	pages, err := p.storage.List(context.Background(), username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}
	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatID, msgNoSavedPages)
	}

	for _, message := range formatPagesList(pages) {
		if err := p.tg.SendMessage(chatID, message); err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}

func (p *Processor) sendStart(chatID int) error {
	return p.tg.SendMessage(chatID, msgHello)
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	u, err := url.Parse(text)
	return err == nil && u.Host != ""
}

func formatPagesList(pages []storage.Page) []string {
	var (
		chunks  []string
		current strings.Builder
	)

	current.WriteString(msgAllLinksHeader)
	current.WriteString("\n\n")

	for i, page := range pages {
		appendMessageLine(&chunks, &current, fmt.Sprintf("%d. %s\n", i+1, page.URL))
	}

	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimRight(current.String(), "\n"))
	}

	return chunks
}

func appendMessageLine(chunks *[]string, current *strings.Builder, line string) {
	if current.Len()+len(line) > telegramMessageLimit && current.Len() > 0 {
		*chunks = append(*chunks, strings.TrimRight(current.String(), "\n"))
		current.Reset()
	}

	for len(line) > telegramMessageLimit {
		if current.Len() > 0 {
			*chunks = append(*chunks, strings.TrimRight(current.String(), "\n"))
			current.Reset()
		}

		*chunks = append(*chunks, line[:telegramMessageLimit])
		line = line[telegramMessageLimit:]
	}

	current.WriteString(line)
}
