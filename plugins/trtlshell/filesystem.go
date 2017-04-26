package trtlshell

import (
	"fmt"
	"strings"
)

/*
	PWD
*/
func (p *TrtlShellPlugin) getPresentWorkingDirectory(user *active) (string, bool) {
  directories := []string{}
  for currentDirectory := user.currentDirectory; currentDirectory != &p.root; currentDirectory = currentDirectory.parent {
    directories = append(directories, currentDirectory.name)
  }

	for i := 0 ; i < len(directories) / 2; i++ {
		swapWith := len(directories) - i - 1
		directories[i], directories[swapWith] = directories[swapWith], directories[i]
	}

  return "/" + strings.Join(directories, "/"), true
}

/*
	CD
*/
func (p *TrtlShellPlugin) changeDirectory(user *active, tokens []string) (string, bool) {
  if len(tokens) != 2 {
    return heckleTheUser("cd")
  }

	currentDirectory := p.getDirectoryAtPath(user, tokens[1])

	if currentDirectory == nil {
		return fmt.Sprintf("'%s' does not exist.", tokens[1]), true
	}

  user.currentDirectory = currentDirectory

  return "", false
}

/*
	LS
*/
func (p *TrtlShellPlugin) listDirectory(user *active, tokens []string) (string, bool) {
	if len(tokens) > 2 {
		return heckleTheUser("ls")
	}

	currentDirectory := user.currentDirectory

	if len(tokens) > 1 {
		currentDirectory = p.getDirectoryAtPath(user, tokens[1])
	}

	if currentDirectory == nil {
		return fmt.Sprintf("'%s' does not exist.", tokens[1]), true
	}

	return currentDirectory.getListing(), true
}

/*
	MKDIR
*/
func (p *TrtlShellPlugin) makeDirectory(user *active, tokens []string) (string, bool) {
  if len(tokens) != 2 {
    return heckleTheUser("mkdir")
  }

  requestedDirectories := strings.Split(tokens[1], "/")

  currentDirectory := user.currentDirectory
	if tokens[1][0] == '/' {
		currentDirectory = &p.root
	}

  for _, requestedDirectory := range requestedDirectories {
		if requestedDirectory == "" {
			continue
		}
		currentDirectory = currentDirectory.createChildDirectoryIfNotPresent(requestedDirectory)
  }

  return "", false
}

/*
	TOUCH
*/
func (p *TrtlShellPlugin) touchFile(user *active, tokens []string) (string, bool) {
  if len(tokens) != 2 {
    return heckleTheUser("touch")
  }

	requestedDirectories := strings.Split(tokens[1], "/")
	currentDirectory := user.currentDirectory

	if len(requestedDirectories) > 1 {
		allButFile := strings.Join(requestedDirectories[:len(requestedDirectories)-1], "/")
		if tokens[1][0] == '/' {
			allButFile = "/" + allButFile
		}
		currentDirectory = p.getDirectoryAtPath(user, allButFile)
	}

	if currentDirectory == nil {
		return fmt.Sprintf("'%s' does not exist.", tokens[1]), true
	}

	filename := requestedDirectories[len(requestedDirectories)-1]

	if _, ok := currentDirectory.childDirectories[filename]; !ok {
		if _, ok := currentDirectory.files[filename]; ok {
			return fmt.Sprintf("'%s' already exists.", filename), true
		}
		currentDirectory.files[filename] = ""
	}

  return "", false
}

func (p *TrtlShellPlugin) catCommand(user *active, tokens []string) (string, bool) {
	if len(tokens) < 2 {
		return heckleTheUser("cat")
	}

	args := strings.Join(tokens[1:], " ")

	if !strings.Contains(args, ">") {
		returnString := ""
		for _, token := range tokens[1:] {
			value, _ := p.getFileContents(user, token)
			returnString += value + "\n"
		}
		return returnString, true
	}

	split := strings.Split(args, ">")
	inputFile := split[0]
	value, _ := p.getFileContents(user, inputFile)
	outputFile := split[len(split)-1]

	if strings.Contains(args, ">>") {
		return p.appendToFile(user, outputFile, value)
	}
	return p.truncateAndWriteFile(user, outputFile, value)
}

func (p *TrtlShellPlugin) echoCommand(user *active, tokens []string) (string, bool) {
	args := strings.Join(tokens[1:], " ")
	if !strings.Contains(args, ">") {
		return strings.Trim(strings.Trim(args, " "), "\""), true
	}
	split := strings.Split(args, ">")
	value := strings.Trim(strings.Trim(split[0], " "), "\"")
	outputFile := split[len(split)-1]

	if strings.Contains(args, ">>") {
		return p.appendToFile(user, outputFile, value)
	}
	return p.truncateAndWriteFile(user, outputFile, value)
}
