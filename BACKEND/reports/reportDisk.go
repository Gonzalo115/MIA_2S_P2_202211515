package reports

import (
	structures "BACKEND/structures"
	utils "BACKEND/utils"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
)

func ReportDisk(mbr *structures.MBR, pathDisk string, path string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}
	// Obtener el nombre base del archivo sin la extensión
	dotFileName, outputImage := utils.GetFileNames(path)

	//Variable en donde haremos el contero de cuanto espacio se utilizado del disco
	espacioUsado := int32(binary.Size(mbr))

	// Definir el contenido DOT con una tabla
	dotContent := fmt.Sprintf(`digraph G {
							node [shape=plaintext]
							labelloc="t"
							label="Reporte DISK"
							fontsize=25;
        					node [shape=plaintext]
        					tabla [label=<
            				<table border="1">
                			<tr>
                    		<td border="2" bgcolor="#3f3f3f"><font color="white"><b>MBR</b></font></td>
							`)

	for _, part := range mbr.Mbr_partitions {
		partType := rune(part.Part_type[0])
		if partType == 'P' {
			dotContent += fmt.Sprintf(`<td border="2" bgcolor="#7e7e7e"><font color="white"><b>PRIMARIA</b><br/>%.4g%% del Disco</font></td>
									 `, tamanoPorcentaje(mbr.Mbr_size, part.Part_size))
			espacioUsado += part.Part_size
		} else if partType == 'E' {
			stringLogica, colspan := concatenarLogica(part.Part_start, part.Part_size, mbr.Mbr_size, pathDisk)

			dotContent += fmt.Sprintf(`<td border="2">
									<table border="1">
									<tr>
									<td colspan="%d" border="1" bgcolor="#7e7e7e"><font color="white"><b>EXTENDIDA</b><br/>%.4g%% del Disco</font></td>
									</tr>
									<tr>`, colspan, tamanoPorcentaje(mbr.Mbr_size, part.Part_size))

			dotContent += stringLogica

			dotContent += fmt.Sprintf(`</tr>
									</table>
            						</td> `)
			espacioUsado += part.Part_size
		} else {
			continue
		}

	}

	if espacioUsado > 0 {
		dotContent += fmt.Sprintf(`<td border="2" bgcolor="#7e7e7e"><font color="white"><b>LIBRE</b><br/>%.4g%% del Disco</font></td>
								`, tamanoPorcentaje(mbr.Mbr_size, mbr.Mbr_size-espacioUsado))
	}

	dotContent += fmt.Sprintf(`</tr>
							</table>>]
							}`)
	// Guardar el contenido DOT en un archivo
	file, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error al crear el archivo DOT: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo DOT: %v", err)
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

func tamanoPorcentaje(tamanoDisk int32, tamanoUsado int32) float32 {

	return (100 * float32(tamanoUsado)) / float32(tamanoDisk)
}

func concatenarLogica(start int32, partSize int32, diskSize int32, pathDisk string) (string, int) {
	var filaLogica string
	var conteo int = 0
	var usado int32 = 0
	offset := start

	for {
		var ebr structures.EBR

		err := ebr.DeserializeEBR(pathDisk, int64(offset))
		if err != nil {
			fmt.Println("Error deserializando el EBR:", err)
		}

		if ebr.Part_next != -1 {
			filaLogica += fmt.Sprintf(`<td border="1" bgcolor="#b0b0b0"><font color="black"><b>EBR</b></font></td>
										<td border="1" bgcolor="#ffffff"><font color="black"><b>LOGICA</b><br/>%.4g%% del Disco <br/>%.4g%% de la Particion</font></td>
									`, tamanoPorcentaje(diskSize, ebr.Part_size), tamanoPorcentaje(partSize, ebr.Part_size))

		} else {
			filaLogica += fmt.Sprintf(`<td border="1" bgcolor="#ffffff"><font color="black"><b>LIBRE</b> <br/>%.4g%% de la Particion</font></td>
									`, tamanoPorcentaje(partSize, partSize-usado))

			conteo += 1
			break
		}
		usado += ebr.SizeEBR() + ebr.Part_size
		conteo += 2
		offset = ebr.Part_next
	}

	return filaLogica, conteo
}
