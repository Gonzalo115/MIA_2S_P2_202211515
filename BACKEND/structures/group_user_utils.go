package structures

import (
	"strings"
)

type GROUP struct {
	Pos   int
	GID   string
	Grupo string
}

type USER struct {
	Pos        int
	UID        string
	Grupo      string
	Usuario    string
	Contrasena string
}

func (sb *SuperBlock) splitGroupsUsers(nd Inode, pathDisk string) ([]GROUP, []USER, error) {

	var listGroup []GROUP
	var listUser []USER
	var contador int = 0

	infoBloque, err := sb.concatenarInfo(nd, pathDisk)
	if err != nil {
		return nil, nil, err
	}

	splitGU := strings.Split(infoBloque, "\n")

	for _, GU := range splitGU {

		if GU == "" {
			break
		}

		split2 := strings.Split(GU, ",")

		if split2[1] == "G" {
			var g GROUP
			g.Pos = contador
			g.GID = split2[0]
			g.Grupo = split2[2]
			listGroup = append(listGroup, g)
			contador += 5 + len(g.GID) + len(g.Grupo)
		} else if split2[1] == "U" {
			var u USER
			u.Pos = contador
			u.UID = split2[0]
			u.Grupo = split2[2]
			u.Usuario = split2[3]
			u.Contrasena = split2[4]
			listUser = append(listUser, u)
			contador += 7 + len(u.UID) + len(u.Grupo) + len(u.Usuario)
		}
	}

	return listGroup, listUser, nil
}

// Conjunto de funcionea para poder concatenar toda la informacion que pueda tener todos bloques
func (sb *SuperBlock) concatenarInfo(nd Inode, pathDisk string) (string, error) {

	var infoBloques string

	for i, blockIndex := range nd.I_block {

		if blockIndex == -1 {
			break
		}

		if i < 12 {
			block := &FileBlock{}
			// Deserializar el bloque
			err := block.Deserialize(pathDisk, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				return "", err
			}
			// Deserializar el contenido del FileBlock
			contenido := strings.Trim(string(block.B_content[:]), "\x00 ")

			// Agrupar en un string toda la informacion de los bloques
			infoBloques += contenido
		} else {
			contenido, err := sb.recursivePointer(pathDisk, blockIndex, i-11) // Donde i representa el AI y 11 el numero de AD dejandonos el nivel del arbol que se be de operar
			if err != nil {
				return "", err
			}
			infoBloques += contenido
		}
	}

	return infoBloques, nil
}

// Recorrer los punteros indirectos de manera recursiva
func (sb *SuperBlock) recursivePointer(pathDisk string, blockIndex int32, nivel int) (string, error) {
	if nivel == 0 {
		fb := &FileBlock{}
		err := fb.Deserialize(pathDisk, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return "", err
		}
		return strings.Trim(string(fb.B_content[:]), "\x00 "), nil
	}

	var infoBloques string
	var pointerBlock PointerBlock

	err := pointerBlock.Deserialize(pathDisk, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
	if err != nil {
		return "", err
	}

	for _, nextBlockIndex := range pointerBlock.P_pointers {
		if nextBlockIndex == -1 {
			break
		}

		contenido, err := sb.recursivePointer(pathDisk, nextBlockIndex, nivel-1)
		if err != nil {
			return "", err
		}
		infoBloques += contenido
	}

	return infoBloques, nil
}

// Conjunto de funcionea para poder concatenar toda la informacion que pueda tener todos bloques
func (sb *SuperBlock) inserInfo(nd *Inode, pathDisk string, info string) error {

	for i, blockIndex := range nd.I_block {

		if i >= 12 {
			in, err := sb.recursiveInserInfo(nd, pathDisk, blockIndex, i-11, info)
			if err != nil {
				return err
			}

			info = in

			if len(info) == 0 {
				return nil // Información insertada completamente, salir de la recursión
			}
			continue
		}

		if blockIndex == -1 {
			usersBlock := &FileBlock{
				B_content: [64]byte{},
			}
			// Serializar el bloque de users.txt
			err := usersBlock.Serialize(pathDisk, int64(sb.S_first_blo))
			if err != nil {
				return err
			}

			// Agregarle un valor artificial a blockIndes
			blockIndex = sb.S_blocks_count

			//Actualizar el inodo
			nd.I_block[i] = sb.S_blocks_count
			err = nd.ActualizarI_mtime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
			if err != nil {
				return err
			}
			//Actualizar todo lo referente al super bloque
			// Actualizar el bitmap de bloques
			err = sb.UpdateBitmapBlock(pathDisk)
			if err != nil {
				return err
			}

			// Actualizamos el superbloque
			sb.S_blocks_count++
			sb.S_free_blocks_count--
			sb.S_first_blo += sb.S_block_size

		}

		if i < 12 {
			block := &FileBlock{}
			// Deserializar el bloque
			err := block.Deserialize(pathDisk, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				return err
			}
			// Deserializar el contenido del FileBlock
			contenido := strings.Trim(string(block.B_content[:]), "\x00 ")

			if len(contenido) == 64 {
				continue
			}

			if len(info) <= (64 - len(contenido)) {
				copy(block.B_content[:], contenido+info)
				block.Serialize(pathDisk, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				nd.I_size = nd.I_size + int32(len(info))
				err = nd.ActualizarI_atime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
				if err != nil {
					return err
				}
				break
			} else {
				disponible := 64 - len(contenido)
				parte1 := info[:disponible]
				info = info[disponible:]
				copy(block.B_content[:], contenido+parte1)
				nd.I_size = nd.I_size + int32(len(parte1))
				err = nd.ActualizarI_atime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
				block.Serialize(pathDisk, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				continue
			}
		}
	}

	return nil
}

func (sb *SuperBlock) recursiveInserInfo(nd *Inode, pathDisk string, blockIndex int32, nivel int, info string) (string, error) {
	// Si estamos en el nivel 0, deserializar el bloque de archivo
	if nivel == 0 {
		fb := &FileBlock{}
		err := fb.Deserialize(pathDisk, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return "", err
		}

		// Leer el contenido actual del bloque
		contenido := strings.Trim(string(fb.B_content[:]), "\x00 ")

		if len(contenido) == 64 {
			return info, nil
		}

		// Verificar si hay espacio suficiente en el bloque actual

		if len(info) <= (64 - len(contenido)) {
			// Espacio suficiente en el bloque actual
			copy(fb.B_content[len(contenido):], info)
			err = fb.Serialize(pathDisk, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))

			if err != nil {
				return "", err
			}
			nd.I_size = nd.I_size + int32(len(info))
			err = nd.ActualizarI_atime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
			if err != nil {
				return "", err
			}
			return "", nil // Se ha insertado toda la información, salir de la recursión
		} else {
			// Parte de la información se guarda en el bloque actual
			disponible := 64 - len(contenido)
			parte1 := info[:disponible]
			info = info[disponible:]
			copy(fb.B_content[len(contenido):], parte1)
			nd.I_size = nd.I_size + int32(len(parte1))
			err = nd.ActualizarI_atime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
			err = fb.Serialize(pathDisk, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				return "", err
			}
			return info, nil
		}
	}

	if nivel == -1 {
		if blockIndex == -1 {
			apuntBlock := PointerBlock{
				P_pointers: [16]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
			}
			// Serializar el bloque de users.txt
			err := apuntBlock.Serialize(pathDisk, int64(sb.S_first_blo))
			if err != nil {
				return "", err
			}

			// Agregarle un valor artificial a blockIndes
			blockIndex = sb.S_blocks_count

			//Actualizar el inodo
			nd.I_block[12] = sb.S_blocks_count
			err = nd.ActualizarI_mtime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
			if err != nil {
				return "", err
			}

			//Actualizar todo lo referente al super bloque
			// Actualizar el bitmap de bloques
			err = sb.UpdateBitmapBlock(pathDisk)
			if err != nil {
				return "", err
			}

			// Actualizamos el superbloque
			sb.S_blocks_count++
			sb.S_free_blocks_count--
			sb.S_first_blo += sb.S_block_size
		}

		// Si estamos en un nivel de punteros, deserializar el bloque de punteros
		pointerBlock := &PointerBlock{}
		err := pointerBlock.Deserialize(pathDisk, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return "", err
		}

		// Procesar cada puntero en el bloque de punteros
		for i, nextBlockIndex := range pointerBlock.P_pointers {
			if nextBlockIndex == -1 {
				usersBlock := &FileBlock{
					B_content: [64]byte{},
				}
				// Serializar el bloque de users.txt
				err := usersBlock.Serialize(pathDisk, int64(sb.S_first_blo))
				if err != nil {
					return "", err
				}

				// Agregarle un valor artificial a blockIndes
				nextBlockIndex = sb.S_blocks_count

				//Actualizar el inodo
				nd.I_block[i] = sb.S_blocks_count
				err = nd.ActualizarI_mtime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
				if err != nil {
					return "", err
				}
				//Actualizar todo lo referente al super bloque
				// Actualizar el bitmap de bloques
				err = sb.UpdateBitmapBlock(pathDisk)
				if err != nil {
					return "", err
				}

				// Actualizamos el superbloque
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size
			}

			// Llamar recursivamente para el siguiente bloque
			info, err := sb.recursiveInserInfo(nd, pathDisk, nextBlockIndex, 0, info)
			if err != nil {
				return "", err
			}

			if len(info) == 0 {
				return "", nil // Información insertada completamente, salir de la recursión
			}
		}

	}

	// Nivel 2: Bloques de punteros dobles
	if nivel == 2 {
		if blockIndex == -1 {
			// Crear un nuevo bloque de punteros dobles
			apuntBlock := &PointerBlock{
				P_pointers: [16]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
			}
			// Serializar el bloque
			err := apuntBlock.Serialize(pathDisk, int64(sb.S_first_blo))
			if err != nil {
				return "", err
			}

			// Actualizar el índice del bloque y el superbloque
			blockIndex = sb.S_blocks_count

			nd.I_block[13] = sb.S_blocks_count
			err = nd.ActualizarI_mtime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
			if err != nil {
				return "", err
			}

			err = sb.UpdateBitmapBlock(pathDisk)
			if err != nil {
				return "", err
			}
			sb.S_blocks_count++
			sb.S_free_blocks_count--
			sb.S_first_blo += sb.S_block_size
		}

		// Deserializar el bloque de punteros dobles
		doublePointerBlock := &PointerBlock{}
		err := doublePointerBlock.Deserialize(pathDisk, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return "", err
		}

		// Procesar cada puntero en el bloque de punteros dobles
		for _, pointerBlockIndex := range doublePointerBlock.P_pointers {

			// Llamar recursivamente para el siguiente bloque
			info, err := sb.recursiveInserInfo(nd, pathDisk, pointerBlockIndex, 1, info)
			if err != nil {
				return "", err
			}

			if len(info) == 0 {
				return "", nil // Información insertada completamente, salir de la recursión
			}
		}

		return info, nil
	}

	// Nivel 3: Bloques de punteros triples
	if nivel == 3 {
		if blockIndex == -1 {
			// Crear un nuevo bloque de punteros triples
			apuntBlock := &PointerBlock{
				P_pointers: [16]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
			}
			// Serializar el bloque
			err := apuntBlock.Serialize(pathDisk, int64(sb.S_first_blo))
			if err != nil {
				return "", err
			}

			// Actualizar el índice del bloque y el superbloque
			blockIndex = sb.S_blocks_count

			nd.I_block[14] = sb.S_blocks_count
			err = nd.ActualizarI_mtime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
			if err != nil {
				return "", err
			}

			err = sb.UpdateBitmapBlock(pathDisk)
			if err != nil {
				return "", err
			}
			sb.S_blocks_count++
			sb.S_free_blocks_count--
			sb.S_first_blo += sb.S_block_size
		}

		// Deserializar el bloque de punteros triples
		triplePointerBlock := &PointerBlock{}
		err := triplePointerBlock.Deserialize(pathDisk, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return "", err
		}

		// Procesar cada puntero en el bloque de punteros triples
		for _, doublePointerBlockIndex := range triplePointerBlock.P_pointers {

			// Llamar recursivamente para el siguiente bloque
			info, err := sb.recursiveInserInfo(nd, pathDisk, doublePointerBlockIndex, 2, info)
			if err != nil {
				return "", err
			}

			if len(info) == 0 {
				return "", nil // Información insertada completamente, salir de la recursión
			}
		}

		return info, nil
	}

	return "", nil
}
