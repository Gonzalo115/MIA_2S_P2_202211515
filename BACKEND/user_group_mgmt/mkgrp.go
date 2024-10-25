package user_group_mgmt

import (
	"BACKEND/global"
	"BACKEND/structures"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type MKGRP struct {
	name string
}

func ParserMkgrp(tokens []string) (string, error) {
	cmd := &MKGRP{}

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

	err := commandMkgrp(cmd)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error al crear un Grupo: ", err))
	}

	return "El grupo se creo exitosamente.", nil
}

func commandMkgrp(mkgrp *MKGRP) error {

	userL, err := global.IsUserRoot()
	if err != nil {
		return err
	}

	if len(mkgrp.name) > 10 {
		return errors.New("El nombre del grupo debe contener como maximo 10 caracteres.")
	}

	//Nueva instancia del superBloque
	superB := &structures.SuperBlock{}
	//Deserializar el superBloque
	superB.Deserialize(userL.PathDisk, int64(userL.Particion.Part_start))

	exist, GID, _, err := superB.ValidateGroup(mkgrp.name, userL.PathDisk)

	if err != nil {
		return err
	}

	if exist {
		return errors.New("El nombre del grupo que se intenta guardar ya esta registrado.")
	}

	newGroup := strconv.Itoa(GID) + ",G," + mkgrp.name + "\n"

	err2 := superB.AddTXT(newGroup, userL.PathDisk)

	if err != nil {
		return err2
	}

	superB.Serialize(userL.PathDisk, int64(userL.Particion.Part_start))

	return nil
}
