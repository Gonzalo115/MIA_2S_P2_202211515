package commands

import (
	global "BACKEND/global"
	structures "BACKEND/structures"
	utils "BACKEND/utils"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// MKFILE estructura que representa el comando mkfile con sus parámetros
type MKFILE struct {
	path string // Ruta del archivo
	r    bool   // Opción recursiva
	size int    // Tamaño del archivo
	cont string // Contenido del archivo
}

// ParserMkfile parsea el comando mkfile y devuelve una instancia de MKFILE
func ParserMkfile(tokens []string) (string, error) {
	cmd := &MKFILE{} // Crea una nueva instancia de MKFILE

	// Itera sobre cada coincidencia encontrada
	for _, match := range tokens {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])
		var value string
		if len(kv) == 2 {
			value = kv[1]
		}

		// Remove quotes from value if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-path":
			// Verifica que el path no esté vacío
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-r":
			// Establece el valor de r a true
			cmd.r = true
		case "-size":
			// Convierte el valor del tamaño a un entero
			size, err := strconv.Atoi(value)
			if err != nil || size < 0 {
				return "", errors.New("el tamaño debe ser un número entero no negativo")
			}
			cmd.size = size
		case "-cont":
			// Verifica que el contenido no esté vacío
			if value == "" {
				return "", errors.New("el contenido no puede estar vacío")
			}
			cmd.cont = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que el parámetro -path haya sido proporcionado
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	// Si no se proporcionó el tamaño, se establece por defecto a 0
	if cmd.size == 0 {
		cmd.size = 0
	}

	// Si no se proporcionó el contenido, se establece por defecto a ""
	if cmd.cont == "" {
		cmd.cont = ""
	}

	// Crear el archivo con los parámetros proporcionados
	err := commandMkfile(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MKFILE: Archivo %s creado correctamente.", cmd.path), nil // Devuelve el comando MKFILE creado
}

// Función ficticia para crear el archivo (debe ser implementada)
func commandMkfile(mkfile *MKFILE) error {

	userL, err := global.IsUserLogeado()
	if err != nil {
		return err
	}

	// Obtener la partición montada
	partitionSuperblock, mountedPartition, partitionPath, err := global.GetMountedPartitionSuperblock(userL.Id)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	mkfile.cont, err = generateContent(mkfile)
	if err != nil {
		return err
	}

	// Crear el archivo
	err = createFile(mkfile.r, mkfile.path, mkfile.size, mkfile.cont, partitionSuperblock, partitionPath, mountedPartition)
	if err != nil {
		err = fmt.Errorf("error al crear el archivo: %w", err)
	}

	return err
}

func generateContent(mkfile *MKFILE) (string, error) {
	// Leer el contenido del archivo

	if mkfile.cont != "" {
		data, err := os.ReadFile(mkfile.cont)
		if err != nil {
			return "", fmt.Errorf("error leyendo el archivo: %w", err)
		}
		return string(data), nil
	}

	if mkfile.size != 0 {
		content := ""
		// Rellenar la cadena hasta cumplir el tamaño requerido
		for len(content) < mkfile.size {
			content += "0123456789"
		}

		// Recortar la cadena al tamaño exacto
		return content, nil
	}
	return "", errors.New("No se porciono contenido valido para archivo")
}

// Funcion para crear un archivo
func createFile(r bool, filePath string, size int, content string, sb *structures.SuperBlock, pathDisk string, mountedPartition *structures.Partition) error {
	fmt.Println("\nCreando archivo:", filePath)

	parentDirs, destDir := utils.GetParentDirectories(filePath)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)

	// Crear el archivo
	err := sb.CreateFile(pathDisk, r, parentDirs, destDir, content)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %w", err)
	}

	// Imprimir inodos y bloques
	sb.PrintInodes(pathDisk)
	sb.PrintBlocks(pathDisk)

	// Serializar el superbloque
	err = sb.Serialize(pathDisk, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}
