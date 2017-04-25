package trtlshell

import (
	"fmt"
	"strings"

	"github.com/velour/catbase/bot"
	"github.com/velour/catbase/bot/msg"
)

type active struct {
  name string
  envVariables map[string]string
  currentDirectory *directory
}

type TrtlShellPlugin struct {
	Bot      bot.Bot
	root directory
  activeSessions map[string]*active
}

func New(bot bot.Bot) *TrtlShellPlugin {
	plugin := &TrtlShellPlugin{
		Bot:      bot,
		root: directory {
      name : "/",
      parent : nil,
      childDirectories : map[string]*directory{},
      files : map[string]string{},
    },
    activeSessions : map[string]*active{},
	}
	return plugin
}


func (p *TrtlShellPlugin) Message(message msg.Message) bool {
  username := message.User.Name
  lowercase := strings.ToLower(message.Body)
  tokens := strings.Fields(lowercase)

  var response string
  var listenToMe bool

  if len(tokens) > 0 {
    if tokens[0] == "trtlshell" {
      response, listenToMe = p.initializeSession(username, tokens)
    } else if user, ok := p.activeSessions[username]; ok {
      if tokens[0] == "exit" {
        response, listenToMe = p.terminateSession(user)
      } else if tokens[0] == "pwd" {
        response, listenToMe = p.getPresentWorkingDirectory(user)
      } else if tokens[0] == "cd" {
        response, listenToMe = p.changeDirectory(user, tokens)
      } else if tokens[0] == "ls" {
        response, listenToMe = p.listDirectory(user, tokens)
      }else if tokens[0] == "mkdir" {
        response, listenToMe = p.makeDirectory(user, tokens)
      }
    }
  }

  if listenToMe {
    p.Bot.SendMessage(message.Channel, response)
  }

  return listenToMe
}

func (p *TrtlShellPlugin) Help(channel string, parts []string) {
	p.Bot.SendMessage(channel, "if you have to ask you'll never know (but type 'exit' to escape)")
}

func (p *TrtlShellPlugin) Event(kind string, message msg.Message) bool {
	return false
}

func (p *TrtlShellPlugin) BotMessage(message msg.Message) bool {
	return false
}

func (p *TrtlShellPlugin) RegisterWeb() *string {
	return nil
}

func (p *TrtlShellPlugin) initializeSession(username string, tokens []string) (string, bool) {
  if _, ok := p.activeSessions[username]; ok {
    return fmt.Sprintf("%s is already logged in. type 'exit' to quit.", username), true
  }

  p.activeSessions[username] = &active{
    name : username,
    envVariables : map[string]string{},
    currentDirectory : p.createUserDirectoryIfNotPresent(username),
  }
	return fmt.Sprintf("%s is now logged in.", username), true
}

func (p *TrtlShellPlugin) terminateSession(user *active) (string, bool) {
  delete(p.activeSessions, user.name)
  return fmt.Sprintf("%s is now logged out.", user.name), true
}
