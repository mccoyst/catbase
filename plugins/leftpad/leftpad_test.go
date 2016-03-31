// © 2013 the CatBase Authors under the WTFPL. See AUTHORS for the list of authors.

package leftpad

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/velour/catbase/bot"
	"github.com/velour/catbase/plugins/counter"
)

func makeMessage(payload string) bot.Message {
	isCmd := strings.HasPrefix(payload, "!")
	if isCmd {
		payload = payload[1:]
	}
	return bot.Message{
		User:    &bot.User{Name: "tester"},
		Channel: "test",
		Body:    payload,
		Command: isCmd,
	}
}

func makePlugin(t *testing.T) (*LeftpadPlugin, *bot.MockBot) {
	mb := bot.NewMockBot()
	counter.New(mb)
	p := New(mb)
	assert.NotNil(t, p)
	return p, mb
}

func TestLeftpad(t *testing.T) {
	p, mb := makePlugin(t)
	p.Message(makeMessage("!leftpad test 8 test"))
	assert.Contains(t, mb.Messages[0], "testtest")
	assert.Len(t, mb.Messages, 1)
}

func TestBadNumber(t *testing.T) {
	p, mb := makePlugin(t)
	p.Message(makeMessage("!leftpad test fuck test"))
	assert.Contains(t, mb.Messages[0], "Invalid")
	assert.Len(t, mb.Messages, 1)
}

func TestNotCommand(t *testing.T) {
	p, mb := makePlugin(t)
	p.Message(makeMessage("leftpad test fuck test"))
	assert.Len(t, mb.Messages, 0)
}

func TestNotPadding(t *testing.T) {
	p, mb := makePlugin(t)
	p.Message(makeMessage("!lololol"))
	assert.Len(t, mb.Messages, 0)
}

func TestHelp(t *testing.T) {
	p, mb := makePlugin(t)
	p.Help("channel", []string{})
	assert.Len(t, mb.Messages, 0)
}

func TestBotMessage(t *testing.T) {
	p, _ := makePlugin(t)
	assert.False(t, p.BotMessage(makeMessage("test")))
}

func TestEvent(t *testing.T) {
	p, _ := makePlugin(t)
	assert.False(t, p.Event("dummy", makeMessage("test")))
}

func TestRegisterWeb(t *testing.T) {
	p, _ := makePlugin(t)
	assert.Nil(t, p.RegisterWeb())
}
