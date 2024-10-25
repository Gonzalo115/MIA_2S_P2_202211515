package global

import (
	"BACKEND/structures"
	"errors"
)

type USERLOGEADO struct {
	grup      string
	uid       string
	Name      string
	Id        string
	Particion structures.Partition
	PathDisk  string
}

var userL = USERLOGEADO{}

// Esta funcion Guardara la informacion del usuario a logear
func SaveUserSession(name string, id string, grup string, uid string) error {
	// Obtener la informacion de la particion montada
	particion, pathDisk, err := GetMountedPartition(id)
	if err != nil {
		return err
	}

	//Guardar la infomacion del usuario a logear
	userL.grup = grup
	userL.uid = uid
	userL.Name = name
	userL.Id = id
	userL.Particion = *particion
	userL.PathDisk = pathDisk

	return nil
}

// Funcion para deslogear a un usuario
func DeleteUserSession() error {

	if ValidateNoLoggedUser() {
		return errors.New("No existe una Sesion actualmente")
	}

	userL.grup = ""
	userL.uid = ""
	userL.Name = ""
	userL.Id = ""
	userL.Particion = structures.Partition{}
	userL.PathDisk = ""

	return nil

}

// Funcion para validar que sea el usuario root
func IsUserRoot() (USERLOGEADO, error) {

	if ValidateNoLoggedUser() {
		return userL, errors.New("No existe una Sesion actualmente")
	}

	if userL.Name != "root" {
		return userL, errors.New("El usuario logeado no es el usuario root")
	}

	return userL, nil
}

// Funcion para validar que sea el usuario root
func IsUserLogeado() (USERLOGEADO, error) {

	if ValidateNoLoggedUser() {
		return userL, errors.New("No existe una Sesion actualmente")
	}

	return userL, nil
}

// Funcion para validar si no hay ningun usuario ya logeado
func ValidateNoLoggedUser() bool {

	if userL.Name == "" {
		return true
	}

	return false
}
