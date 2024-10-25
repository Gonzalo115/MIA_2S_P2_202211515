package structures

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
)

type MBR struct {
	Mbr_size           int32
	Mbr_creation_date  float32
	Mbr_disk_signature int32
	Mbr_disk_fit       [1]byte
	Mbr_partitions     [4]Partition
}

// SerializeMBR escribe la estructura MBR al inicio de un archivo binario
func (mbr *MBR) Serialize(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	err = binary.Write(file, binary.LittleEndian, mbr)
	if err != nil {
		return err
	}

	return nil
}

// DeserializeMBR lee la estructura MBR desde el inicio de un archivo binario
func (mbr *MBR) Deserialize(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	mbrSize := binary.Size(mbr)
	if mbrSize <= 0 {
		return fmt.Errorf("invalid MBR size: %d", mbrSize)
	}

	buffer := make([]byte, mbrSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, mbr)
	if err != nil {
		return err
	}

	return nil
}

// Método para obtener la primera partición disponible
func (mbr *MBR) GetFirstAvailablePartition(sizeBytes int) (*Partition, int, int, bool) {
	offset := binary.Size(mbr)

	for i := 0; i < len(mbr.Mbr_partitions); i++ {
		if mbr.Mbr_partitions[i].Part_type[0] == byte('0') && mbr.Mbr_partitions[i].Part_start == -1 {

			// Si el tamaño es mayor entonces no se puede guardar la particion
			if int(sizeBytes) > (int(mbr.Mbr_size) - offset) {
				return nil, -1, -1, true
			}

			return &mbr.Mbr_partitions[i], offset, i, false
		} else {
			offset += int(mbr.Mbr_partitions[i].Part_size)
		}
	}
	return nil, -1, -1, false
}

// Obterner el ebr que se encuentra dispobible osea el ultimo para poder asinarle un particion logica
func (mbr *MBR) GetFirstAvailableEBR(path string, sizeBytes int) (*EBR, int32, int32, bool, bool) {

	var offset int32
	var Part_size int32
	var usado int32 = 0

	for i := range mbr.Mbr_partitions {
		if strings.EqualFold(strings.Trim(string(mbr.Mbr_partitions[i].Part_type[:]), "\x00 "), "E") {
			offset = mbr.Mbr_partitions[i].Part_start
			Part_size = mbr.Mbr_partitions[i].Part_size
		}
	}

	for {
		var ebr EBR
		ebr.DeserializeEBR(path, int64(offset))
		usado = usado + ebr.Part_size + ebr.SizeEBR()

		if ebr.Part_next == -1 {

			newEBR := false

			if int32(sizeBytes)+ebr.SizeEBR() >= (Part_size - usado) {
				return nil, -1, -1, false, true
			}

			if Part_size > (usado + int32(ebr.SizeEBR())) {
				newEBR = true
			}

			return &ebr, offset, ebr.SizeEBR(), newEBR, false
		}
		offset = ebr.Part_next
	}
}

// Buscar si el nombre de particion que se desea guardar ya existe
func (mbr *MBR) SearchNameMatches(name string, path string) bool {
	for i := range mbr.Mbr_partitions {
		partitionName := strings.Trim(string(mbr.Mbr_partitions[i].Part_name[:]), "\x00 ")
		inputName := strings.Trim(name, "\x00 ")
		if strings.EqualFold(partitionName, inputName) {
			return true
		}

		if strings.EqualFold(strings.Trim(string(mbr.Mbr_partitions[i].Part_type[:]), "\x00 "), "E") {
			if SearchNameMatchesLogical(inputName, int64(mbr.Mbr_partitions[i].Part_start), path) {
				return true
			}

		}
	}
	return false
}

// Buscar en la particion extendida si existe una parte logica tiene el mismo nombre
func SearchNameMatchesLogical(inputName string, offset int64, path string) bool {

	for {
		var ebr EBR

		ebr.DeserializeEBR(path, offset)

		if ebr.Part_start == -1 {
			return false
		}

		if strings.EqualFold(strings.Trim(inputName, "\x00 "), strings.Trim(string(ebr.Part_name[:]), "\x00 ")) {
			return true
		}

		if ebr.Part_next == -1 {
			return false
		}

		offset = int64(ebr.Part_next)
	}
}

// Buscar si en el disco ya existe una particion extendida
func (mbr *MBR) ContainsExtendedPartition() bool {
	for i := range mbr.Mbr_partitions {
		if strings.EqualFold(strings.Trim(string(mbr.Mbr_partitions[i].Part_type[:]), "\x00 "), "E") {
			return true
		}
	}
	return false
}

// Método para obtener una partición por nombre
func (mbr *MBR) GetPartitionByName(name string) (*Partition, int, int, bool) {

	var id int

	for i, partition := range mbr.Mbr_partitions {
		partitionName := strings.Trim(string(partition.Part_name[:]), "\x00 ")
		inputName := strings.Trim(name, "\x00 ")

		if strings.EqualFold(strings.Trim(string(partition.Part_type[:]), "\x00 "), "P") {
			id = id + 1
		}

		if strings.EqualFold(partitionName, inputName) {

			if strings.EqualFold(strings.Trim(string(partition.Part_type[:]), "\x00 "), "E") {
				return nil, -1, -1, true
			}

			if strings.EqualFold(strings.Trim(string(partition.Part_status[:]), "\x00 "), "1") {
				return nil, -1, -10, false
			}
			return &partition, i, id, false
		}
	}
	return nil, -1, -1, false
}

// Función para obtener una partición por ID
func (mbr *MBR) GetPartitionByID(id string) (*Partition, error) {
	for i := 0; i < len(mbr.Mbr_partitions); i++ {
		partitionID := strings.Trim(string(mbr.Mbr_partitions[i].Part_id[:]), "\x00 ")
		inputID := strings.Trim(id, "\x00 ")
		if strings.EqualFold(partitionID, inputID) {
			return &mbr.Mbr_partitions[i], nil
		}
	}
	return nil, errors.New("partición no encontrada")
}
