package reports

import (
	structures "BACKEND/structures"
	utils "BACKEND/utils"
	"fmt"
	"os"
	"os/exec"
)

// ReportInode genera un reporte de un inodo y lo guarda en la ruta especificada
func ReportBlock(superblock *structures.SuperBlock, diskPath string, path string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	// Obtener el nombre base del archivo sin la extensión
	dotFileName, outputImage := utils.GetFileNames(path)
	fmt.Printf(dotFileName, outputImage)

	println("avanzo")

	// Iniciar el contenido DOT
	dotContent := `digraph G {
        node [shape=plaintext]
    `

	// Iterar sobre cada inodo
	for i := int32(0); i < superblock.S_inodes_count; i++ {
		inode := &structures.Inode{}
		// Deserializar el inodo
		err := inode.Deserialize(diskPath, int64(superblock.S_inode_start+(i*superblock.S_inode_size)))
		if err != nil {
			return err
		}

		// Iterar sobre cada bloque del inodo (apuntadores)
		for _, blockIndex := range inode.I_block {
			// Si el bloque no existe, salir
			if blockIndex == -1 {
				break
			}
			// Si el inodo es de tipo carpeta
			if inode.I_type[0] == '0' {
				var block structures.FolderBlock
				// Deserializar el bloque
				err := block.Deserialize(diskPath, int64(superblock.S_block_start+(blockIndex*superblock.S_block_size))) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				//Concatenar aqui

				dotContent += fmt.Sprintf(`
				inode%d [label=<
					<table border="0" cellborder="1" cellspacing="0">
						<tr>
							<td colspan="2">BLOQUE DE CARPETA %d </td>
						</tr>
						<tr>
							<td>Nombre</td>
							<td>Inodo</td>
						</tr>
				`, int(blockIndex), int(blockIndex)) // i es el índice del inodo

				dotContent += block.Agregar()
				dotContent += `</table>>];`

				continue

				// Si el inodo es de tipo archivo
			} else if inode.I_type[0] == '1' {
				var block structures.FileBlock
				// Deserializar el bloque
				err := block.Deserialize(diskPath, int64(superblock.S_block_start+(blockIndex*superblock.S_block_size))) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				dotContent += fmt.Sprintf(`
				inode%d [label=<
					<table border="0" cellborder="1" cellspacing="0">
						<tr>
							<td>BLOQUE DE ARCHIVO %d </td>
						</tr>
				`, int(blockIndex), int(blockIndex)) // i es el índice del inodo
				dotContent += "<tr><td>"
				dotContent += block.AgregarF()
				dotContent += "</td></tr>"
				dotContent += `</table>>];`

				continue
			}

		}
	}

	// Cerrar el contenido DOT
	dotContent += "}"

	println(dotContent)

	// Crear el archivo DOT
	dotFile, err := os.Create(dotFileName)
	if err != nil {
		return err
	}
	defer dotFile.Close()

	// Escribir el contenido DOT en el archivo
	_, err = dotFile.WriteString(dotContent)
	if err != nil {
		return err
	}

	// Generar la imagen con Graphviz
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
