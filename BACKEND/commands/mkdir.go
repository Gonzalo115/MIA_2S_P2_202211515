package commands

import (
	global "BACKEND/global"
	structures "BACKEND/structures"
	utils "BACKEND/utils"
	"errors"
	"fmt"
	"strings"
)

// MKDIR estructura que representa el comando mkdir con sus parámetros
type MKDIR struct {
	path string // Path del directorio
	p    bool   // Opción -p (crea directorios padres si no existen)
}

/*
   mkdir -p -path=/home/user/docs/usac
   mkdir -path="/home/mis documentos/archivos clases"
*/

func ParserMkdir(tokens []string) (string, error) {
	cmd := &MKDIR{} // Crea una nueva instancia de MKDIR

	// Itera sobre cada coincidencia encontrada
	for _, match := range tokens {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])

		// Switch para manejar diferentes parámetros
		switch key {
		case "-path":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			// Remove quotes from value if present
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.path = value
		case "-p":
			if len(kv) > 2 {
				return "", errors.New(fmt.Sprintf("Se estan pasando mas pareametros que  los necesarios %s", match))
			}
			cmd.p = true
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que el parámetro -path haya sido proporcionado
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	// Aquí se puede agregar la lógica para ejecutar el comando mkdir con los parámetros proporcionados
	err := commandMkdir(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MKDIR: Directorio %s creado correctamente."), nil // Devuelve el comando MKDIR creado
}

func commandMkdir(mkdir *MKDIR) error {

	userL, err := global.IsUserLogeado()
	if err != nil {
		return err
	}

	// Obtener la partición montada
	partitionSuperblock, mountedPartition, partitionPath, err := global.GetMountedPartitionSuperblock(userL.Id)
	if err != nil {
		return errors.New(fmt.Sprintf("error al obtener la partición montada: ", err))
	}

	// Crear el directorio
	err = createDirectory(mkdir.path, mkdir.p, partitionSuperblock, partitionPath, mountedPartition)
	if err != nil {
		err = fmt.Errorf("error al crear el directorio: %w", err)
	}

	return err
}

func createDirectory(dirPath string, p bool, sb *structures.SuperBlock, partitionPath string, mountedPartition *structures.Partition) error {
	//fmt.Println("\nCreando directorio:", dirPath)

	parentDirs, destDir := utils.GetParentDirectories(dirPath)

	// Crear el directorio segun el path proporcionado
	err := sb.CreateFolder(partitionPath, p, parentDirs, destDir)
	if err != nil {
		return fmt.Errorf("error al crear el directorio: %w", err)
	}

	fmt.Println("Despues que se rompara")
	// Imprimir inodos y bloques
	sb.PrintInodes(partitionPath)
	sb.PrintBlocks(partitionPath)

	// Serializar el superbloque
	err = sb.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}
