package trtlshell

import (
	"testing"
  "strings"

  "github.com/stretchr/testify/assert"
  "github.com/velour/catbase/bot/msg"
  "github.com/velour/catbase/bot"
  "github.com/velour/catbase/bot/user"
)

func makeMessage(payload string) msg.Message {
	isCmd := strings.HasPrefix(payload, "!")
	if isCmd {
		payload = payload[1:]
	}
	return msg.Message{
		User:    &user.User{Name: "tester"},
		Channel: "test",
		Body:    payload,
		Command: isCmd,
	}
}

func TestIgnore(t *testing.T) {
  mb := bot.NewMockBot()
	c := New(mb)
	assert.NotNil(t, c)
  c.Message(makeMessage("exit"))
  res := assert.Len(t, mb.Messages, 0)
  assert.True(t, res)
}

func TestLogin(t *testing.T) {
  mb := bot.NewMockBot()
	c := New(mb)
	assert.NotNil(t, c)
	c.Message(makeMessage("!trtlshell"))
  res := assert.Len(t, mb.Messages, 1)
  assert.True(t, res)
  assert.Contains(t, mb.Messages[0], "tester is now logged in.")
}

func TestLogout(t *testing.T) {
  mb := bot.NewMockBot()
	c := New(mb)
	assert.NotNil(t, c)
	c.Message(makeMessage("!trtlshell"))
  c.Message(makeMessage("exit"))
  res := assert.Len(t, mb.Messages, 2)
  assert.True(t, res)
  assert.Contains(t, mb.Messages[0], "tester is now logged in.")
  assert.Contains(t, mb.Messages[1], "tester is now logged out.")
}

func TestDoubleLogin(t *testing.T) {
  mb := bot.NewMockBot()
	c := New(mb)
	assert.NotNil(t, c)
	c.Message(makeMessage("!trtlshell"))
  c.Message(makeMessage("!trtlshell"))
  res := assert.Len(t, mb.Messages, 2)
  assert.True(t, res)
  assert.Contains(t, mb.Messages[0], "tester is now logged in.")
  assert.Contains(t, mb.Messages[1], "tester is already logged in.")
}

func TestPWD(t *testing.T) {
  mb := bot.NewMockBot()
	c := New(mb)
	assert.NotNil(t, c)
	c.Message(makeMessage("!trtlshell"))
  c.Message(makeMessage("pwd"))
  c.Message(makeMessage("exit"))
  res := assert.Len(t, mb.Messages, 3)
  assert.True(t, res)
  assert.Contains(t, mb.Messages[0], "tester is now logged in.")
  assert.Contains(t, mb.Messages[1], "/usr/tester")
  assert.Contains(t, mb.Messages[2], "tester is now logged out.")
}

func TestCD(t *testing.T) {
  mb := bot.NewMockBot()
	c := New(mb)
	assert.NotNil(t, c)
	c.Message(makeMessage("!trtlshell"))
  c.Message(makeMessage("cd .."))
  c.Message(makeMessage("pwd"))
  c.Message(makeMessage("cd .."))
  c.Message(makeMessage("pwd"))
  c.Message(makeMessage("cd .."))
  c.Message(makeMessage("pwd"))
  c.Message(makeMessage("cd usr/tester"))
  c.Message(makeMessage("pwd"))
  c.Message(makeMessage("exit"))
  res := assert.Len(t, mb.Messages, 6)
  assert.True(t, res)
  assert.Contains(t, mb.Messages[0], "tester is now logged in.")
  assert.Contains(t, mb.Messages[1], "/usr")
  assert.Contains(t, mb.Messages[2], "/")
  assert.Contains(t, mb.Messages[3], "/")
  assert.Contains(t, mb.Messages[4], "/usr/tester")
  assert.Contains(t, mb.Messages[5], "tester is now logged out.")
}

func TestCDFail(t *testing.T) {
  mb := bot.NewMockBot()
	c := New(mb)
	assert.NotNil(t, c)
	c.Message(makeMessage("!trtlshell"))
  c.Message(makeMessage("cd not valid arguments"))
  c.Message(makeMessage("cd notadirectory"))
  c.Message(makeMessage("exit"))
  res := assert.Len(t, mb.Messages, 4)
  assert.True(t, res)
  assert.Contains(t, mb.Messages[0], "tester is now logged in.")
  assert.Contains(t, mb.Messages[1], "really? you don't know how to use cd")
  assert.Contains(t, mb.Messages[2], "directory 'notadirectory' does not exist.")
  assert.Contains(t, mb.Messages[3], "tester is now logged out.")
}
