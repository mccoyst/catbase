package trtlshell

import (
	"fmt"
	"strings"
)

type directory struct {
  name string
  parent *directory
  childDirectories map[string]*directory
  files map[string]string
}

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

func (p *TrtlShellPlugin) changeDirectory(user *active, tokens []string) (string, bool) {
  if len(tokens) != 2 {
    return "really? you don't know how to use cd", true
  }

	currentDirectory := p.getDirectoryAtPath(user, tokens[1])

	if currentDirectory == nil {
		return fmt.Sprintf("'%s' does not exist.", tokens[1]), true
	}

  user.currentDirectory = currentDirectory

  return "", false
}

func (p *TrtlShellPlugin) listDirectory(user *active, tokens []string) (string, bool) {
	if len(tokens) > 2 {
		return "really? you don't know how to use ls", true
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

func (p *TrtlShellPlugin) getDirectoryAtPath(user *active, path string) *directory {
	requestedDirectories := strings.Split(path, "/")

	currentDirectory := user.currentDirectory
	if path[0] == '/' {
		currentDirectory = &p.root
	} else if path[0] == '~' {
		currentDirectory = p.root.createChildDirectoryIfNotPresent("home")
	}

	for _, requestedDirectory := range requestedDirectories {
		if requestedDirectory == "" || requestedDirectory == "." {
			continue
		}

		if requestedDirectory == ".." {
			if currentDirectory.parent != nil {
				currentDirectory = currentDirectory.parent
			}
		} else {
			if dir, ok := currentDirectory.childDirectories[requestedDirectory]; ok {
				currentDirectory = dir
			} else {
				return nil
			}
		}
	}
	return currentDirectory
}

func (p *TrtlShellPlugin) makeDirectory(user *active, tokens []string) (string, bool) {
  if len(tokens) != 2 {
    return "really? you don't know how to use mkdir", true
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

func (p *TrtlShellPlugin) createUserDirectoryIfNotPresent(username string) *directory {
  home := p.root.createChildDirectoryIfNotPresent("home")
  return home.createChildDirectoryIfNotPresent(username)
}

func (parent *directory) createChildDirectoryIfNotPresent(newDirectoryName string) *directory {
  if existing, ok := parent.childDirectories[newDirectoryName]; ok {
    //hmmm this could be an error but for now we'll ignore it
    return existing
  }
  parent.childDirectories[newDirectoryName] = &directory {
    name : newDirectoryName,
    parent : parent,
    childDirectories : map[string]*directory{},
    files : map[string]string{},
  }
  return parent.childDirectories[newDirectoryName]
}

func (dir *directory) getListing() string {
	contents := []string{".", ".."}

	for _, child := range dir.childDirectories {
		contents = append(contents, child.name)
	}

	for filename, _ := range dir.files {
		contents = append(contents, filename)
	}

	return strings.Join(contents, "\n")
}
