package user_group_mgmt

import (
	"errors"
	"fmt"
	"strings"
)

type RMUSR struct {
	user string
}

func ParserRmusr(tokens []string) (string, error) {
	cmd := &RMUSR{}

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
		default:
			return "", errors.New(fmt.Sprintf("Parámetro desconocido: %s", key))
		}
	}

	if cmd.user == "" {
		return "", errors.New("Faltan parámetros requeridos: -name")
	}

	err := commandRmusr(cmd)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error al eliminar un usuario: ", err))
	}

	return "Se ha eliminado el usuario exitosamente.", nil
}

func commandRmusr(rmusr *RMUSR) error {
	fmt.Println(rmusr.user)
	return nil
}
