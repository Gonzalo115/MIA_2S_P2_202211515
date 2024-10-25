package user_group_mgmt

import (
	"errors"
	"fmt"
	"strings"
)

type CHGRP struct {
	user string
	grp  string
}

func ParserChgrp(tokens []string) (string, error) {
	cmd := &CHGRP{}

	for _, match := range tokens {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", errors.New(fmt.Sprintf("formato de parámetro inválido: %s", match))
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
		case "-grp":
			if value == "" {
				return "", errors.New("El grp no puede estar vacio")
			}
			cmd.grp = value
		default:
			return "", errors.New(fmt.Sprintf("parámetro desconocido: %s", key))
		}
	}

	if cmd.user == "" {
		return "", errors.New("faltan parámetros requeridos: -user")
	}

	if cmd.grp == "" {
		return "", errors.New("faltan parámetros requeridos: -pass")
	}

	err := commandChgrp(cmd)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error al cambiar de grupo a el usario: ", err))
	}

	return "Se ha cambiado de grupo ha el usuario", nil
}

func commandChgrp(chgrp *CHGRP) error {
	fmt.Println(chgrp.user)
	fmt.Println(chgrp.grp)
	return nil
}
