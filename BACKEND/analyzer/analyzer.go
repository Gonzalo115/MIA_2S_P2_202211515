package analyzer

import (
	commands "BACKEND/commands"
	usergroupmgmt "BACKEND/user_group_mgmt"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Analyzer analiza el comando de entrada y ejecuta la acción correspondiente
func Analyzer(input string) (string, error) {

	re := regexp.MustCompile(`(?:[^\s=]+="[^"]*"|[^\s]+)`)
	tokens := re.FindAllString(input, -1)

	if len(tokens) == 0 {
		return "", errors.New("no se proporcionó ningún comando")
	}

	comando := strings.ToLower(tokens[0])

	switch comando {
	case "mkdisk":
		return commands.ParserMkdisk(tokens[1:])
	case "rmdisk":
		return commands.ParserRmdisk(tokens[1:])
	case "fdisk":
		return commands.ParserFdisk(tokens[1:])
	case "mount":
		return commands.ParserMount(tokens[1:])
	case "mkfs":
		return commands.ParserMkfs(tokens[1:])
	case "mkdir":
		return commands.ParserMkdir(tokens[1:])
	case "mkfile":
		return commands.ParserMkfile(tokens[1:])
	case "login":
		return usergroupmgmt.ParserLogin(tokens[1:])
	case "logout":
		return usergroupmgmt.ParserLogout(tokens[1:])
	case "mkgrp":
		return usergroupmgmt.ParserMkgrp(tokens[1:])
	case "rmgrp":
		return usergroupmgmt.ParserRmgrp(tokens[1:])
	case "mkusr":
		return usergroupmgmt.ParserMkusr(tokens[1:])
	case "rmusr":
		return usergroupmgmt.ParserRmusr(tokens[1:])
	case "chgrp":
		return usergroupmgmt.ParserChgrp(tokens[1:])
	case "rep":
		return commands.ParserRep(tokens[1:])
	case "execute":
		return "", CommandExecute(tokens[1:])
	case "clear":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return "", errors.New("no se pudo limpiar la terminal")
		}
		return "", nil
	default:
		return "", fmt.Errorf("comando desconocido: %s", tokens[0])
	}
}
