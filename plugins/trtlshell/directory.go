package trtlshell

import "strings"

type directory struct {
	name             string
	parent           *directory
	childDirectories map[string]*directory
	files            map[string]string
}

func (p *TrtlShellPlugin) getDirectoryAtPath(user *active, path string) *directory {
	requestedDirectories := strings.Split(path, "/")

	currentDirectory := user.currentDirectory
	if path[0] == '/' {
		currentDirectory = &p.root
	} else if path[0] == '~' {
		if len(path) == 1 || path[1] == '/' {
			currentDirectory = p.getDirectoryAtPath(user, "/home/"+user.name)
		} else {
			possibleUsername := requestedDirectories[0][1:]
			if _, ok := p.users[possibleUsername]; !ok {
				return nil
			}
			currentDirectory = p.getDirectoryAtPath(user, "/home/"+possibleUsername)
		}
		requestedDirectories = requestedDirectories[1:]
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

func (p *TrtlShellPlugin) createUserDirectoryIfNotPresent(username string) *directory {
	home := p.root.createChildDirectoryIfNotPresent("home")
	return home.createChildDirectoryIfNotPresent(username)
}

func (parent *directory) createChildDirectoryIfNotPresent(newDirectoryName string) *directory {
	if _, ok := parent.files[newDirectoryName]; ok {
		return nil
	}

	if existing, ok := parent.childDirectories[newDirectoryName]; ok {
		//hmmm this could be an error but for now we'll ignore it
		return existing
	}
	parent.childDirectories[newDirectoryName] = &directory{
		name:             newDirectoryName,
		parent:           parent,
		childDirectories: map[string]*directory{},
		files:            map[string]string{},
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
