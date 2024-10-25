package user_group_mgmt

import (
	"errors"
	"fmt"
	"strings"
)

type RMGRP struct {
	name string
}

func ParserRmgrp(tokens []string) (string, error) {
	cmd := &RMGRP{}

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
		case "-name":
			if value == "" {
				return "", errors.New("el name no puede estar vacío")
			}
			cmd.name = value
		default:
			return "", errors.New(fmt.Sprintf("parámetro desconocido: %s", key))
		}
	}

	if cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -name")
	}

	err := commandRmgrp(cmd)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error al eliminar un grupo"))
	}

	return "Se ha eliminado el grupo existosamente", nil
}

func commandRmgrp(rmgrp *RMGRP) error {
	fmt.Println(rmgrp.name)
	return nil
}
