package commands

import (
	global "BACKEND/global"
	reports "BACKEND/reports"
	"errors"
	"fmt"
	"strings"
)

// REP estructura que representa el comando rep con sus parámetros
type REP struct {
	id           string
	path         string
	name         string
	path_file_ls string
}

func ParserRep(tokens []string) (string, error) {
	cmd := &REP{}

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
		case "-id":
			if value == "" {
				return "", errors.New("el id no puede estar vacío")
			}
			cmd.id = value
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-name":
			validNames := []string{"mbr", "disk", "inode", "block", "bm_inode", "bm_bloc", "sb", "file", "ls"}
			if !contains(validNames, value) {
				return "", errors.New("nombre inválido, debe ser uno de los siguientes: mbr, disk, inode, block, bm_inode, bm_block, sb, file, ls")
			}
			cmd.name = value
		case "-path_file_ls":
			cmd.path_file_ls = value
		default:
			return "", errors.New(fmt.Sprintf("parámetro desconocido: %s", key))
		}
	}

	if cmd.id == "" || cmd.path == "" || cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -id, -path, -name")
	}

	mens, err := commandRep(cmd)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error al crear Reporte: ", err))
	}

	return mens, nil
}

// Función auxiliar para verificar si un valor está en una lista
func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

// Ejemplo de función commandRep (debe ser implementada)
func commandRep(rep *REP) (string, error) {
	mountedMbr, mountedSb, pathDisk, err := global.GetMountedPartitionRep(rep.id)

	mountedSb.Print()

	if err != nil {
		return "", err
	}

	switch rep.name {
	case "mbr":
		err = reports.ReportMBR(mountedMbr, pathDisk, rep.path)
		if err != nil {
			return "", err
		}
		return "Se ha generado el reporte del MBR", nil
	case "disk":
		err = reports.ReportDisk(mountedMbr, pathDisk, rep.path)
		if err != nil {
			return "", err
		}
		return "Se ha generado el reporte del Particiones del Disk", nil
	case "inode":
		err = reports.ReportInode(mountedSb, pathDisk, rep.path)
		if err != nil {
			return "", err
		}
		return "SE han generado el reporte de inodos", nil
	case "block":
		err = reports.ReportBlock(mountedSb, pathDisk, rep.path)
		if err != nil {
			println("aqui trono")
			return "", err
		}
		return "Se han generado el reporte de bloques", nil
	case "bm_inode":
		err = reports.ReportBMInode(mountedSb, pathDisk, rep.path)
		if err != nil {
			return "", err
		}
		return "Se ha generado el reporte de mapa de bits de los indodos", nil
	case "bm_bloc":
		err = reports.ReportBMIbloc(mountedSb, pathDisk, rep.path)
		if err != nil {
			return "", err
		}
		return "Se ha generado el reporte de mapa de bits de los bloques", nil
	case "sb":
		err = reports.ReportSB(mountedSb, pathDisk, rep.path)
		if err != nil {
			return "", err
		}
		return "Se ha generado el reporte de Super Bloque", nil
	}

	return "", nil
}
