package reports

import (
	structures "BACKEND/structures"
	utils "BACKEND/utils"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func ReportSB(sb *structures.SuperBlock, pathDisk string, path string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}
	// Obtener el nombre base del archivo sin la extensión
	dotFileName, outputImage := utils.GetFileNames(path)

	mountTime := time.Unix(int64(sb.S_mtime), 0)
	// Convertir el tiempo de desmontaje a una fecha
	unmountTime := time.Unix(int64(sb.S_umtime), 0)

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
					<td border="2" bgcolor="#3f3f3f"><font color="white"><b>sb</b></font></td>
					<td border="2" bgcolor="#3f3f3f"><font color="white"><b>Value</b></font></td>
				</tr>
				<tr>
					<td>sb_nombre_hd</td>
					<td>%s</td>
				</tr>
				<tr>
					<td>S_filesystem_type</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_inodes_count</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_blocks_count</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_free_inodes_count</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_free_blocks_count</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_mtime</td>
					<td>%s</td>
				</tr>
				<tr>
					<td>S_umtime</td>
					<td>%s</td>
				</tr>
				<tr>
					<td>S_mnt_count</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_magic</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_inode_size</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_block_size</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_first_ino</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_first_blo</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_bm_inode_start</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_bm_block_start</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_inode_start</td>
					<td>%d</td>
				</tr>
				<tr>
					<td>S_block_start</td>
					<td>%d</td>
				</tr>
			</table>>]
	}`,
		pathDisk,
		sb.S_filesystem_type,
		sb.S_inodes_count,
		sb.S_blocks_count,
		sb.S_free_inodes_count,
		sb.S_free_blocks_count,
		mountTime.Format(time.RFC3339),
		unmountTime.Format(time.RFC3339),
		sb.S_mnt_count,
		sb.S_magic,
		sb.S_inode_size,
		sb.S_block_size,
		sb.S_first_ino,
		sb.S_first_blo,
		sb.S_bm_inode_start,
		sb.S_bm_block_start,
		sb.S_inode_start,
		sb.S_block_start,
	)

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
	return nil
}
