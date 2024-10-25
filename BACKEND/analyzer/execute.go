package analyzer

//Gonzalo Fernando Perez Cazun - 202211515
import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func CommandExecute(tokens []string) error {
	path := tokens[0]

	tokens2 := strings.Split(path, "=")

	direccion := tokens2[1]

	err := ReadFileLines(direccion)

	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %w", err)
	}

	return nil
}

func ReadFileLines(filePath string) error {

	cleanedPath := strings.Trim(filePath, `"`)
	// Abrir el archivo
	file, err := os.Open(cleanedPath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %w", err)
	}
	defer file.Close()

	// Crear un nuevo escáner para leer el archivo línea por línea
	scanner := bufio.NewScanner(file)

	// Iterar sobre cada línea del archivo
	for scanner.Scan() {
		line := scanner.Text()
		// Procesar cada línea (en este caso, simplemente imprimirla)
		fmt.Print(">>>> ")
		fmt.Println(line)
		_, err := Analyzer(line)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
	}

	// Verificar si ocurrió un error durante la lectura
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error al leer el archivo: %w", err)
	}

	return nil
}
