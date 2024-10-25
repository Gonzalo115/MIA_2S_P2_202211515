package structures

type Partition struct {
	Part_status      [1]byte
	Part_type        [1]byte
	Part_fit         [1]byte
	Part_start       int32
	Part_size        int32
	Part_name        [16]byte
	Part_correlative int32
	Part_id          [4]byte
}

/*
	0: Inactiva
	1: Montada
*/

// Crear una partición con los parámetros proporcionados
func (p *Partition) CreatePartition(partStart, partSize int, partType, partFit, partName string) {
	// Asignar status de la partición
	p.Part_status[0] = '0' // El valor '0' indica que la parcion esta Inactiva

	p.Part_start = int32(partStart) // Inicio de la particion

	p.Part_size = int32(partSize) // Tamaño de la particion

	if len(partType) > 0 {
		p.Part_type[0] = partType[0] //Tipo de asignancion
	}

	if len(partFit) > 0 {
		p.Part_fit[0] = partFit[0] // Ajuste de la particion
	}

	copy(p.Part_name[:], partName) // Nombre de la particion

	p.Part_correlative = 0
}

func (p *Partition) MountPartition(correlative int, id string) error {
	p.Part_correlative = int32(correlative) + 1
	copy(p.Part_id[:], id)
	return nil
}
