package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type EBR struct {
	Part_mount [1]byte  // La particion esta montada "0 No montada" y "1 montada"
	Part_fit   [1]byte  // Tipo de ajuste de la particion B, F o W
	Part_start int32    // Inicio de la particion Logica
	Part_size  int32    // Tamaño total de la particion en bytes
	Part_next  int32    // Donde se encuentra el siguiente EBR con -1 inidica que no hay otro ebr
	Part_name  [16]byte // Nombre de la particion
}

// SerializeEBR escribe la estructura EBR en una posición específica de un archivo binario
func (ebr *EBR) SerializeEBR(path string, offset int64) error {

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mueve el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, os.SEEK_SET)
	if err != nil {
		return fmt.Errorf("error al mover el puntero de archivo: %w", err)
	}

	// Escribe la estructura EBR en la posición actual del puntero
	err = binary.Write(file, binary.LittleEndian, ebr)
	if err != nil {
		return fmt.Errorf("error al escribir datos en el archivo: %w", err)
	}

	return nil
}

// Deserializeebr lee la estructura ebr desde el inicio de un archivo binario
func (ebr *EBR) DeserializeEBR(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mueve el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, os.SEEK_SET)
	if err != nil {
		return fmt.Errorf("error al mover el puntero de archivo: %w", err)
	}

	ebrSize := binary.Size(ebr)
	if ebrSize <= 0 {
		return fmt.Errorf("invalid EBR size: %d", ebrSize)
	}

	buffer := make([]byte, ebrSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, ebr)
	if err != nil {
		return err
	}

	return nil
}

// Crear una partición con los parámetros proporcionados
func (ebr *EBR) CreateEBR(partStart, partSize int, partNext int, partType, partFit, partName string) {
	// Asignar status de la partición
	ebr.Part_mount[0] = '0' // El valor '0' indica que la parcion esta Inactiva

	if len(partFit) > 0 {
		ebr.Part_fit[0] = partFit[0] // Ajuste de la particion
	}

	ebr.Part_start = int32(partStart) // Inicio de la particion

	ebr.Part_size = int32(partSize) // Tamaño de la particion

	ebr.Part_next = int32(partNext) // Indicar donde empieza el proximo ebr

	copy(ebr.Part_name[:], partName) // Nombre de la particion
}

func (ebr *EBR) SizeEBR() int32 {
	return int32(binary.Size(ebr))
}
