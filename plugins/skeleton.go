package plugins

import "bitbucket.org/phlyingpenguin/godeepintir/bot"

// This is a skeleton plugin to serve as an example and quick copy/paste for new plugins.

type SkeletonPlugin struct {
	Bot *bot.Bot
}

// NewSkeletonPlugin creates a new SkeletonPlugin with the Plugin interface
func NewSkeletonPlugin(bot *bot.Bot) *SkeletonPlugin {
	return &SkeletonPlugin{
		Bot: bot,
	}
}

// Message responds to the bot hook on recieving messages.
// This function returns true if the plugin responds in a meaningful way to the users message.
// Otherwise, the function returns false and the bot continues execution of other plugins.
func (p *SkeletonPlugin) Message(message bot.Message) bool {
	// This bot does not reply to anything
	return false
}

// LoadData imports any configuration data into the plugin. This is not strictly necessary other
// than the fact that the Plugin interface demands it exist. This may be deprecated at a later
// date.
func (p *SkeletonPlugin) LoadData() {
	// This bot has no data to load
}

// Help responds to help requests. Every plugin must implement a help function.
func (p *SkeletonPlugin) Help(channel string, parts []string) {
	p.Bot.SendMessage(channel, "Sorry, Skeleton does not do a goddamn thing.")
}