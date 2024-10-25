package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type RMDISK struct {
	path string
}

func ParserRmdisk(tokens []string) (string, error) {
	cmd := &RMDISK{}

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
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		default:
			return "", errors.New(fmt.Sprintf("parámetro desconocido: %s", key))
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	err := commandRmdisk(cmd)
	if err != nil {
		return "", err
	}

	return "El disco fue eliminado existosamente", nil
}

func commandRmdisk(rmdisk *RMDISK) error {
	err := os.Remove(rmdisk.path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error al eliminar el disco: %v", err))
	}
	return nil
}
