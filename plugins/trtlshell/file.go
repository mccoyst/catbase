package trtlshell

import (
	"fmt"
	"strings"
)

func (p *TrtlShellPlugin) getFileContents(user *active, filepath string) (string, bool) {
	currentDirectory, filename := p.getFileAndDirectory(user, filepath)
	if contents, ok := currentDirectory.files[filename]; ok {
		return contents, true
	} else {
		return fmt.Sprintf("'%s' does not exist or is not a file.", filepath), true
	}
}


func (p *TrtlShellPlugin) truncateAndWriteFile(user *active, filepath, contents string) (string, bool) {
	currentDirectory, filename := p.getFileAndDirectory(user, filepath)
	currentDirectory.files[filename] = contents
	return "", false
}

func (p *TrtlShellPlugin) appendToFile(user *active, filepath, contents string) (string, bool) {
	currentDirectory, filename := p.getFileAndDirectory(user, filepath)
	currentDirectory.files[filename] += "\n" + contents
	return "", false
}

func (p *TrtlShellPlugin) getFileAndDirectory(user *active, filepath string) (*directory, string) {
	var currentDirectory *directory
	var filename string

	filepath = strings.Trim(filepath, " ")

	if !strings.Contains(filepath, "/") {
		currentDirectory = user.currentDirectory
		filename = filepath
	} else {
		requestedDirectories := strings.Split(filepath, "/")

		allButFile := strings.Join(requestedDirectories[:len(requestedDirectories)-1], "/")
		if filepath[0] == '/' {
			allButFile = "/" + allButFile
		}

		currentDirectory = p.getDirectoryAtPath(user, allButFile)
		filename = requestedDirectories[len(requestedDirectories)-1]
	}
	return currentDirectory, filename
}
