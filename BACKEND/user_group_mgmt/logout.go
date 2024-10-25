package user_group_mgmt

import (
	global "BACKEND/global"
	"errors"
	"fmt"
)

func ParserLogout(tokens []string) (string, error) {

	if len(tokens) != 0 {
		return "", errors.New("Error LOGOUT: Se estan recibiendo mas parametros de los necesarios")
	}

	err := CommandLogout()
	if err != nil {
		fmt.Println("Error al cerrar sesion: ", err)
	}
	return "Se ha cerrado sesion exitosamente.", nil
}

func CommandLogout() error {

	err := global.DeleteUserSession()
	if err != nil {
		return err
	}

	return nil
}
