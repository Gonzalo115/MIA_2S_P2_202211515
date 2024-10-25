package commands

import (
	structures "BACKEND/structures"
	utils "BACKEND/utils"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type MKDISK struct {
	size int
	unit string
	fit  string
	path string
}

func ParserMkdisk(tokens []string) (string, error) {
	cmd := &MKDISK{}

	for _, match := range tokens {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", errors.New("Formato de parámetro inválido al crear disco")
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
			if value != "K" && value != "M" {
				return "", errors.New("la unidad debe ser K o M")
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
		default:
			println("error")
			return "", errors.New(fmt.Sprintf("parámetro desconocido: %s", key))
		}
	}

	if cmd.size == 0 {
		return "", errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	if cmd.unit == "" {
		cmd.unit = "M"
	}

	if cmd.fit == "" {
		cmd.fit = "FF"
	}

	err := commandMkdisk(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return "El Disco Fue creado Existosamente", nil
}

func commandMkdisk(mkdisk *MKDISK) error {
	sizeBytes, err := utils.ConvertToBytes(mkdisk.size, mkdisk.unit)
	if err != nil {
		fmt.Println("Error converting size:", err)
		return err
	}

	err = createDisk(mkdisk, sizeBytes)
	if err != nil {
		fmt.Println("Error creating disk:", err)
		return err
	}

	err = createMBR(mkdisk, sizeBytes)
	if err != nil {
		fmt.Println("Error creating MBR:", err)
		return err
	}

	return nil
}

func createDisk(mkdisk *MKDISK, sizeBytes int) error {
	err := os.MkdirAll(filepath.Dir(mkdisk.path), os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directories:", err)
		return err
	}

	file, err := os.Create(mkdisk.path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	buffer := make([]byte, 1024*1024)
	for sizeBytes > 0 {
		writeSize := len(buffer)
		if sizeBytes < writeSize {
			writeSize = sizeBytes
		}
		if _, err := file.Write(buffer[:writeSize]); err != nil {
			return err
		}
		sizeBytes -= writeSize
	}
	return nil
}

func createMBR(mkdisk *MKDISK, sizeBytes int) error {
	mbr := &structures.MBR{
		Mbr_size:           int32(sizeBytes),
		Mbr_creation_date:  float32(time.Now().Unix()),
		Mbr_disk_signature: rand.Int31(),
		Mbr_disk_fit:       [1]byte{mkdisk.fit[0]},
		Mbr_partitions: [4]structures.Partition{
			{Part_status: [1]byte{'9'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: -1, Part_id: [4]byte{'0'}},
			{Part_status: [1]byte{'9'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: -1, Part_id: [4]byte{'0'}},
			{Part_status: [1]byte{'9'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: -1, Part_id: [4]byte{'0'}},
			{Part_status: [1]byte{'9'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: -1, Part_id: [4]byte{'0'}},
		},
	}

	err := mbr.Serialize(mkdisk.path)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return nil
}
