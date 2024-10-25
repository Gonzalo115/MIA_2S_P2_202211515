package commands

import (
	global "BACKEND/global"
	structures "BACKEND/structures"
	utils "BACKEND/utils"
	"errors"
	"fmt"
	"strings"
)

type MOUNT struct {
	path string
	name string
}

// CommandMount parsea el comando mount y devuelve una instancia de MOUNT
func ParserMount(tokens []string) (string, error) {
	cmd := &MOUNT{}

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
		case "-path":
			if value == "" {
				return "", errors.New("El path no puede estar vacío")
			}
			cmd.path = value
		case "-name":
			if value == "" {
				return "", errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		default:
			return "", errors.New(fmt.Sprintf("parámetro desconocido: %s", key))
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -name")
	}

	// Montamos la partición
	err := commandMount(cmd)
	if err != nil {
		return "", err
	}

	text := "Se ha montado la particion exitosamente\n"
	text += global.StringMont()

	return text, nil
}

func commandMount(mount *MOUNT) error {
	var mbr structures.MBR

	// Deserializar la estructura MBR desde un archivo binario
	err := mbr.Deserialize(mount.path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error deserializando el MBR, ", err))
	}

	// Buscar la partición con el nombre especificado
	partition, indexPartition, id, partitionExtend := mbr.GetPartitionByName(mount.name)

	if id == -10 {
		return fmt.Errorf("La particion ya esta montada")
	}

	if partitionExtend {
		return errors.New("la particion con nombre especificado es extendida")
	}

	if partition == nil {
		return errors.New(fmt.Sprintf("La partición con nombre %s no existe", mount.name))
	}

	// Generar un id único para la partición
	idPartition, err := GenerateIdPartition(mount, id)
	if err != nil {
		return errors.New(fmt.Sprintf("Error generando el id de partición, ", err))
	}

	//  Guardar la partición montada en la lista de montajes globales
	global.MountedPartitions[idPartition] = mount.path

	// Modificamos la partición para indicar que está montada
	partition.MountPartition(indexPartition, idPartition)

	partition.Part_status[0] = '1'

	// Guardar la partición modificada en el MBR
	mbr.Mbr_partitions[indexPartition] = *partition

	// Serializar la estructura MBR en el archivo binario
	err = mbr.Serialize(mount.path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error serializando el MBR, ", err))
	}

	return nil
}

func GenerateIdPartition(mount *MOUNT, indexPartition int) (string, error) {
	// Asignar una letra a la partición
	letter, err := utils.GetLetter(mount.path)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error obteniendo la letra, ", err))
	}

	// Crear id de partición
	idPartition := fmt.Sprintf("%s%d%s", global.Carnet, indexPartition, letter)

	return idPartition, nil
}
