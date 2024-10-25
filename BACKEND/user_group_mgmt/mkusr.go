package user_group_mgmt

import (
	global "BACKEND/global"
	"BACKEND/structures"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type MKUSR struct {
	user string
	pass string
	grp  string
}

func ParserMkusr(tokens []string) (string, error) {
	cmd := &MKUSR{}

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
		case "-grp":
			if value == "" {
				return "", errors.New("El grp no puede estar vacio")
			}
			cmd.grp = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.user == "" {
		return "", errors.New("faltan parámetros requeridos: -user")
	}

	if cmd.pass == "" {
		return "", errors.New("faltan parámetros requeridos: -pass")
	}

	if cmd.grp == "" {
		return "", errors.New("faltan parámetros requeridos: -grp")
	}

	err := commandMkusr(cmd)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error al crear Usuario: ", err))
	}

	return "Se ha creado el usuario exitosamente", nil
}

func commandMkusr(mkusr *MKUSR) error {
	userL, err := global.IsUserRoot()
	if err != nil {
		return err
	}

	if len(mkusr.user) > 10 {
		return errors.New("El nombre del usuuario debe contener como maximo 10 caracteres.")
	}

	if len(mkusr.pass) > 10 {
		return errors.New("El password del usuario debe contener como maximo 10 caracteres.")
	}

	if len(mkusr.grp) > 10 {
		return errors.New("El grupo del usuario debe contener como maximo 10 caracteres.")
	}

	//Nueva instancia del superBloque
	superB := &structures.SuperBlock{}
	//Deserializar el superBloque
	superB.Deserialize(userL.PathDisk, int64(userL.Particion.Part_start))

	UID, err := superB.ValidateUser(mkusr.user, mkusr.grp, userL.PathDisk)

	if err != nil {
		return err
	}

	newUser := strconv.Itoa(UID) + ",U," + mkusr.grp + "," + mkusr.user + "," + mkusr.pass + "\n"

	err2 := superB.AddTXT(newUser, userL.PathDisk)

	if err != nil {
		return err2
	}

	superB.Serialize(userL.PathDisk, int64(userL.Particion.Part_start))

	return nil
}
