package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/luqasn/travis2slack/config"
	"github.com/luqasn/travis2slack/slack"
	"github.com/luqasn/travis2slack/templates"
	"github.com/luqasn/travis2slack/travis"
	log "github.com/sirupsen/logrus"
)

func RespondWithError(w http.ResponseWriter, m string) {
	w.WriteHeader(401)
	w.Write([]byte(m))
}

func RespondWithSuccess(w http.ResponseWriter, m string) {
	w.WriteHeader(200)
	w.Write([]byte(m))
}

type templateWithFilter struct {
	filter   string
	template string
	channel  string
}

func (bot *Bot) DeployHandler(w http.ResponseWriter, r *http.Request) {
	err := bot.travis.VerifySignature(r)
	if err != nil && !bot.config.DisableVerification {
		RespondWithError(w, errors.New("unauthorized payload").Error())
		return
	}

	payload := r.FormValue("payload")

	var jsonData interface{}
	json.Unmarshal([]byte(payload), &jsonData)

	filterMap := map[string]*templateWithFilter{}
	for key, values := range r.URL.Query() {
		re := regexp.MustCompile(`(filter|template|channel)(\[(.+)\])?`)
		res := re.FindAllStringSubmatch(key, -1)
		for i := range res {
			t := res[i][1]
			n := res[i][3]
			m, ok := filterMap[n]
			if !ok {
				m = &templateWithFilter{}
				filterMap[n] = m
			}
			if t == "filter" {
				m.filter = values[0]
			} else if t == "template" {
				m.template = values[0]
			} else if t == "channel" {
				m.channel = values[0]
			}
		}
	}

	if len(filterMap) == 0 {
		RespondWithError(w, errors.New("'channel' arg is required").Error())
		return
	}

	keys := []string{}
	for k := range filterMap {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, k := range keys {
		o := filterMap[k]
		templateString := o.template
		filter := o.filter
		channel := o.channel

		if filter == "" {
			filter = bot.config.DefaultFilter
		}

		if templateString == "" {
			templateString = bot.config.DefaultTemplate
		}

		if channel == "" {
			if c, ok := filterMap[""]; ok && c.channel != "" {
				channel = c.channel
			} else {
				RespondWithError(w, fmt.Sprintf("no channel specified for %q and no default channel found", k))
				return
			}
		}

		channelTpl, err := templates.NewTemplate("channel", channel, map[string]string{})
		if err != nil {
			RespondWithError(w, fmt.Sprintf("Creating %q failed: #%v", filter, err))
			return
		}

		channel, err = channelTpl.Render(jsonData)
		if err != nil {
			RespondWithError(w, fmt.Sprintf("Rendering %q failed: #%v", templateString, err))
			return
		}

		filterTpl, err := templates.NewTemplate("filters", filter, bot.config.Filters)

		if err != nil {
			RespondWithError(w, fmt.Sprintf("Creating %q failed: #%v", filter, err))
			return
		}

		filterRes, err := filterTpl.Render(jsonData)
		if err != nil {
			RespondWithError(w, fmt.Sprintf("Rendering %q failed: #%v", templateString, err))
			return
		}

		matches, err := templates.FilterResultToBool(filterRes)
		if err != nil {
			RespondWithError(w, fmt.Sprintf("Filtering %q failed: #%v", filter, err))
			return
		}

		if !matches {
			continue
		}

		subTemplates := map[string]string{}
		for key, tpl := range bot.config.Templates {
			subTemplates[key] = tpl.Message
		}

		tpl, err := templates.NewTemplate("main", templateString, subTemplates)

		if err != nil {
			RespondWithError(w, fmt.Sprintf("Creating %q failed: #%v", templateString, err))
			return
		}

		x, err := tpl.Render(jsonData)
		if err != nil {
			RespondWithError(w, fmt.Sprintf("Rendering %q failed: #%v", templateString, err))
			return
		}

		if !strings.HasPrefix(channel, "#") {
			channel = "#" + channel
		}

		slacker := slack.NewSlack(bot.config.Slack.OAuthAccessToken)
		slacker.PostMessage(channel, x)
		RespondWithSuccess(w, fmt.Sprintf("Sent %q to %q", x, channel))
		return
	}
	RespondWithSuccess(w, "Filter does not match")
}

type Bot struct {
	config config.Config
	travis travis.Travis
}

func NewBot() Bot {
	cfg := config.LoadConfig()
	return Bot{
		config: cfg,
		travis: travis.NewTravis(cfg.Travis.PublicKeyURL),
	}
}

func main() {
	bot := NewBot()
	http.HandleFunc("/", bot.DeployHandler)
	log.Fatal(http.ListenAndServe(bot.config.HTTP.ListenAddress, nil))
}
