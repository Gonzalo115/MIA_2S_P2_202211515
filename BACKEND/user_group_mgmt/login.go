package user_group_mgmt

import (
	global "BACKEND/global"
	"BACKEND/structures"
	"errors"
	"fmt"
	"strings"
)

type LOGIN struct {
	user string
	pass string
	id   string
}

func ParserLogin(tokens []string) (string, error) {
	cmd := &LOGIN{}

	for _, match := range tokens {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-user":
			if value == "" {
				return "", errors.New("El user no puede estar vacío")
			}
			cmd.user = value
		case "-pass":
			if value == "" {
				return "", errors.New("El pass no puede estar vacio")
			}
			cmd.pass = value
		case "-id":
			if value == "" {
				return "", errors.New("El id no puede estar vacio")
			}
			cmd.id = value
		default:
			return "", errors.New(fmt.Sprintf("parámetro desconocido: %s", key))
		}
	}

	if cmd.user == "" {
		return "", errors.New("faltan parámetros requeridos: -user")
	}

	if cmd.pass == "" {
		return "", errors.New("faltan parámetros requeridos: -pass")
	}

	if cmd.id == "" {
		return "", errors.New("faltan parámetros requeridos: -id")
	}

	err := commandLogin(cmd)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error en Login: ", err))
	}

	return "Se ha iniciado secision exitosamente", nil
}

func commandLogin(login *LOGIN) error {

	// Validar que no existe ya un sesicion inciada
	logeado := global.ValidateNoLoggedUser()
	if !logeado {
		return errors.New("Ya hay unsuario logeado")
	}

	particion, pathDisk, err := global.GetMountedPartition(login.id)
	if err != nil {
		return err
	}

	superB := &structures.SuperBlock{}

	superB.Deserialize(pathDisk, int64(particion.Part_start))

	usuarioExist, user, err := superB.ValidateCredentials(login.user, login.pass, pathDisk)
	if err != nil {
		return err
	}

	if usuarioExist {
		err := global.SaveUserSession(login.user, login.id, user.Grupo, user.UID)
		if err != nil {
			return err
		}
	} else {
		return errors.New("Las credenciales proporcionadas no coiciden en el sistema")
	}

	return nil
}
