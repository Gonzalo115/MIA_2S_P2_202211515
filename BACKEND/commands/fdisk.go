package commands

import (
	structures "BACKEND/structures"
	utils "BACKEND/utils"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type FDISK struct {
	size int
	unit string
	fit  string
	path string
	typ  string
	name string
}

func ParserFdisk(tokens []string) (string, error) {
	cmd := &FDISK{}

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
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return "", errors.New("el tamaño debe ser un número entero positivo")
			}
			cmd.size = size
		case "-unit":
			value = strings.ToUpper(value)
			if value != "K" && value != "M" && value != "B" {
				return "", errors.New("la unidad debe ser B, K o M")
			}
			cmd.unit = strings.ToUpper(value)
		case "-fit":
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return "", errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.fit = value
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-type":
			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				return "", errors.New("el tipo debe ser P, E o L")
			}
			cmd.typ = value
		case "-name":

			if value == "" {
				return "", errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.size == 0 {
		return "", errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -name")
	}

	if cmd.unit == "" {
		cmd.unit = "K"
	}

	if cmd.fit == "" {
		cmd.fit = "WF"
	}

	if cmd.typ == "" {
		cmd.typ = "P"
	}

	err := commandFdisk(cmd)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error en Crear una particion: ", err))
	}

	return "La particion fue creada existosamente", nil
}

func commandFdisk(fdisk *FDISK) error {
	sizeBytes, err := utils.ConvertToBytes(fdisk.size, fdisk.unit)
	if err != nil {
		return err
	}

	if fdisk.typ == "P" {
		err = createPrimaryPartition(fdisk, sizeBytes)
		if err != nil {
			return err
		}
	} else if fdisk.typ == "E" {
		err = createExtendedPartition(fdisk, sizeBytes)
		if err != nil {
			return err
		}

	} else if fdisk.typ == "L" {
		err = createLogicalPartition(fdisk, sizeBytes)
		if err != nil {
			return err
		}
	}

	return nil
}

func createPrimaryPartition(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR

	// Deserializar del archivo el mbr
	err := mbr.Deserialize(fdisk.path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error deserializando el MBR, ", err))
	}

	// Buscar si el nombre que queremos colocar ya existe
	if mbr.SearchNameMatches(fdisk.name, fdisk.path) {
		return errors.New(fmt.Sprintf("El nombre de la particion ya axite"))
	}

	// Buscar en que lugar colocaremos nuestra nueva particion
	availablePartition, startPartition, indexPartition, insuficienteEspacio := mbr.GetFirstAvailablePartition(sizeBytes)

	if insuficienteEspacio {
		return errors.New(fmt.Sprintf("Insuficiente espacio en el disco para almacenar la particion."))
	}

	if availablePartition == nil {
		return errors.New(fmt.Sprintf("No hay particiones disponibles"))
	}

	// Reiscribir nuestra particion
	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)

	// Colocar la partición en el MBR
	if availablePartition != nil {
		mbr.Mbr_partitions[indexPartition] = *availablePartition
	}

	// Serializar el MBR en el archivo binario
	err = mbr.Serialize(fdisk.path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error:", err))
	}

	return nil
}

func createExtendedPartition(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR

	// Deserializar del archivo el mbr
	err := mbr.Deserialize(fdisk.path)
	if err != nil {
		return err
	}

	// Buscar si ya existe una particion Extendida
	if mbr.ContainsExtendedPartition() {
		return errors.New(fmt.Sprintf("Ya exite una particion extendida en el disco"))
	}

	// Buscar si el nombre que queremos colocar ya existe
	if mbr.SearchNameMatches(fdisk.name, fdisk.path) {
		return errors.New(fmt.Sprintf("El nombre de la particion ya axite"))
	}

	// Buscar en que lugar colocaremos nuestra nueva particion
	availablePartition, startPartition, indexPartition, insuficienteEspacio := mbr.GetFirstAvailablePartition(sizeBytes)

	if insuficienteEspacio {
		return errors.New(fmt.Sprintf("Insuficiente espacio en el disco para almacenar la particion."))
	}

	if availablePartition == nil {
		fmt.Sprintf("No hay particiones disponibles.")
		return errors.New(fmt.Sprintf("No hay particiones disponibles."))
	}

	// Reiscribir nuestra particion
	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)

	// Colocar la partición en el MBR
	if availablePartition != nil {
		mbr.Mbr_partitions[indexPartition] = *availablePartition
	}

	// Serializar el MBR en el archivo binario
	err = mbr.Serialize(fdisk.path)
	if err != nil {
		return err
	}

	// Crear el EBR
	err = createEBR(fdisk.path, availablePartition.Part_start)
	if err != nil {
		return err
	}

	return nil
}

func createLogicalPartition(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR

	// Deserializar del archivo el mbr
	err := mbr.Deserialize(fdisk.path)
	if err != nil {
		return err
	}

	// Buscar si ya existe una particion Extendida
	if !mbr.ContainsExtendedPartition() {
		return errors.New(fmt.Sprintf("No existe una particion Extendida en el disco"))
	}

	// Buscar si el nombre que queremos colocar ya existe
	if mbr.SearchNameMatches(fdisk.name, fdisk.path) {
		return errors.New(fmt.Sprintf("El nombre de la particion ya axite"))
	}

	availableEBR, start, SizeEBR, newEBR, insuficienteEspacio := mbr.GetFirstAvailableEBR(fdisk.path, sizeBytes)
	if insuficienteEspacio {
		return errors.New(fmt.Sprintf("No hay suficiente Espacio para la paticion logica"))
	}

	partStart := int(start) + int(SizeEBR)

	var partNext int = -1

	if newEBR {
		partNext = partStart + sizeBytes
	}

	availableEBR.CreateEBR(partStart, sizeBytes, partNext, fdisk.typ, fdisk.fit, fdisk.name)

	err = availableEBR.SerializeEBR(fdisk.path, int64(start))
	if err != nil {
		return err
	}

	// Si ya hay cabida para poder escribir un nuevo ebr no inicializado ya no se escribi
	if newEBR {
		// Crear el EBR
		err = createEBR(fdisk.path, availableEBR.Part_next)
		if err != nil {
			return err
		}
	}

	return nil
}

func createEBR(path string, offset int32) error {
	ebr := &structures.EBR{
		Part_mount: [1]byte{'X'},
		Part_fit:   [1]byte{'X'},
		Part_start: -1,
		Part_size:  -1,
		Part_next:  -1,
		Part_name:  [16]byte{'X'},
	}

	err := ebr.SerializeEBR(path, int64(offset))
	if err != nil {
		return errors.New(fmt.Sprintf("Error al crear el EBR", err))
	}
	return nil
}
