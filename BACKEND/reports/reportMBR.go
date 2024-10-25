package reports

import (
	structures "BACKEND/structures"
	utils "BACKEND/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ReportMBR genera un reporte del MBR y lo guarda en la ruta especificada
func ReportMBR(mbr *structures.MBR, pathDisk string, path string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	// Obtener el nombre base del archivo sin la extensión
	dotFileName, outputImage := utils.GetFileNames(path)

	// Definir el contenido DOT con una tabla
	dotContent := fmt.Sprintf(`digraph G {
		node [shape=plaintext]
		tabla [label=<
			<table border="0" cellborder="1" cellspacing="0" cellpadding="4" color="black">
				<tr><td colspan="2" bgcolor="#008000"><font color="white"><b>REPORTE MBR</b></font></td></tr>
				<tr><td bgcolor="#ddffce">mbr_tamano</td><td bgcolor="#ddffce">%d</td></tr>
				<tr><td bgcolor="#ddffce">mrb_fecha_creacion</td><td bgcolor="#ddffce">%s</td></tr>
				<tr><td bgcolor="#ddffce">mbr_disk_signature</td><td bgcolor="#ddffce">%d</td></tr>
		`, mbr.Mbr_size, time.Unix(int64(mbr.Mbr_creation_date), 0), mbr.Mbr_disk_signature)

	// Agregar las particiones a la tabla
	for i, part := range mbr.Mbr_partitions {

		partStatus := rune(part.Part_status[0])
		partType := rune(part.Part_type[0])
		partFit := rune(part.Part_fit[0])
		partName := strings.TrimRight(string(part.Part_name[:]), "\x00")

		dotContent += fmt.Sprintf(`
				<tr><td colspan="2" bgcolor="#fff000"><font color="black"><b> PARTICIÓN %d </b></font></td></tr>
				<tr><td bgcolor="#fbf6a9">part_status</td><td bgcolor="#fbf6a9">%c</td></tr>
				<tr><td bgcolor="#fbf6a9">part_type</td><td bgcolor="#fbf6a9">%c</td></tr>
				<tr><td bgcolor="#fbf6a9">part_fit</td><td bgcolor="#fbf6a9">%c</td></tr>
				<tr><td bgcolor="#fbf6a9">part_start</td><td bgcolor="#fbf6a9">%d</td></tr>
				<tr><td bgcolor="#fbf6a9">part_size</td><td bgcolor="#fbf6a9">%d</td></tr>
				<tr><td bgcolor="#fbf6a9">part_name</td><td bgcolor="#fbf6a9">%s</td></tr>
			`, i+1, partStatus, partType, partFit, part.Part_start, part.Part_size, partName)
		if strings.EqualFold(strings.Trim(string(mbr.Mbr_partitions[i].Part_type[:]), "\x00 "), "E") {
			offset := part.Part_start
			for {
				var ebr structures.EBR

				err := ebr.DeserializeEBR(pathDisk, int64(offset))
				if err != nil {
					fmt.Println("Error deserializando el EBR:", err)
					return err
				}

				logicalStatus := rune(ebr.Part_mount[0])
				logicalfit := rune(ebr.Part_fit[0])
				logicalName := strings.TrimRight(string(ebr.Part_name[:]), "\x00")

				dotContent += fmt.Sprintf(`
				<tr><td colspan="2" bgcolor="#ff2121"><font color="white"><b> PARTICIÓN LOGICA </b></font></td></tr>
				<tr><td bgcolor="#fed6d6">part_status</td><td bgcolor="#fed6d6">%c</td></tr>
				<tr><td bgcolor="#fed6d6">part_next</td><td bgcolor="#fed6d6">%d</td></tr>
				<tr><td bgcolor="#fed6d6">part_fit</td><td bgcolor="#fed6d6">%c</td></tr>
				<tr><td bgcolor="#fed6d6">part_start</td><td bgcolor="#fed6d6">%d</td></tr>
				<tr><td bgcolor="#fed6d6">part_size</td><td bgcolor="#fed6d6">%d</td></tr>
				<tr><td bgcolor="#fed6d6">part_name</td><td bgcolor="#fed6d6">%s</td></tr>
				`, logicalStatus, ebr.Part_next, logicalfit, ebr.Part_start, ebr.Part_size, logicalName)

				if ebr.Part_next == -1 {
					break
				}
				offset = ebr.Part_next
			}

		}
	}
	dotContent += "</table>>] }"

	// Guardar el contenido DOT en un archivo
	file, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo: %v", err)
	}

	// Ejecutar el comando Graphviz para generar la imagen
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar el comando Graphviz: %v", err)
	}

	fmt.Println("Imagen de la tabla generada:", outputImage)
	return nil
}
