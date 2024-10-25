package structures

import (
	"BACKEND/utils"
	"errors"
	"strings"
	"time"
)

// createFolderInInode crea una carpeta en un inodo específico
func (sb *SuperBlock) createFolderInInode(path string, p bool, inodeIndex int32, parentsDir []string, destDir string) (bool, error) {
	// Crear un nuevo inodo
	inode := &Inode{}
	// Deserializar el inodo
	err := inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return false, err
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] == '1' {
		return false, nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for i, blockIndex := range inode.I_block {

		// Si el bloque no existe, salir
		if blockIndex == -1 {

			if p || len(parentsDir) == 0 {
				var carpeta [12]byte

				if len(parentsDir) == 0 {
					copy(carpeta[:], destDir)
				} else {
					parentDir, err := utils.First(parentsDir)
					if err != nil {
						return false, err
					}
					var parentBytes [12]byte
					copy(parentBytes[:], []byte(parentDir))
					carpeta = parentBytes
				}

				// Creamos el bloque del Inodo Raíz
				Block := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: 0},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				Block.B_content[2].B_name = carpeta
				Block.B_content[2].B_inodo = sb.S_inodes_count

				// Serializar el bloque de users.txt
				err = Block.Serialize(path, int64(sb.S_first_blo))
				if err != nil {
					return false, err
				}

				//Actualizar el inodo
				inode.I_block[i] = sb.S_blocks_count
				err = inode.ActualizarI_mtime(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
				if err != nil {
					return false, err
				}

				// Creamos el inodo newBloque
				inodeNew := &Inode{
					I_uid:   1,
					I_gid:   1,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				// Serializar
				err = inodeNew.Serialize(path, int64(sb.S_first_ino))
				if err != nil {
					return false, err
				}

				// Agregarle un valor artificial a blockIndes
				blockIndex = sb.S_blocks_count

				//Actualizar todo lo referente al super bloque

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInode(path)
				if err != nil {
					return false, err
				}

				// Actualizar el bitmap de bloques
				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return false, err
				}

				// Actualizamos el superbloque
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size

				sb.S_inodes_count++
				sb.S_free_inodes_count--
				sb.S_first_ino += sb.S_inode_size

				if len(parentsDir) == 0 {
					// Crear el bloque de la carpeta
					folderBlock := &FolderBlock{
						B_content: [4]FolderContent{
							{B_name: [12]byte{'.'}, B_inodo: 0},
							{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
							{B_name: [12]byte{'-'}, B_inodo: -1},
							{B_name: [12]byte{'-'}, B_inodo: -1},
						},
					}

					// Serializar el bloque de la carpeta
					err = folderBlock.Serialize(path, int64(sb.S_first_blo))
					if err != nil {
						return false, err
					}

					// Actualizar el bitmap de bloques
					err = sb.UpdateBitmapBlock(path)
					if err != nil {
						return false, err
					}

					inodeNew.I_block[0] = sb.S_blocks_count
					err = inodeNew.Serialize(path, int64(sb.S_first_ino))
					if err != nil {
						return false, err
					}

					// Actualizar el superbloque
					sb.S_blocks_count++
					sb.S_free_blocks_count--
					sb.S_first_blo += sb.S_block_size
					return true, nil
				}

			} else {
				return false, errors.New("No se encuentra en el directorio las carpetas padre y no se porpociono el parametro p")
			}

		}

		// Crear un nuevo bloque de carpeta
		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return false, err
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.B_content[indexContent]

			// Sí las carpetas padre no están vacías debereamos buscar la carpeta padre más cercana
			if len(parentsDir) != 0 {
				//fmt.Println("---------ESTOY  VISITANDO--------")

				// Si el contenido está vacío, salir
				if content.B_inodo == -1 {
					if !p {
						return false, errors.New("No se encuentra en el directorio las carpetas padre y no se porpociono el parametro p")
					}

					// Obtenemos la carpeta padre más cercana
					parentDir, err := utils.First(parentsDir)
					if err != nil {
						return false, err
					}

					var parentBytes [12]byte
					copy(parentBytes[:], []byte(parentDir))

					block.B_content[indexContent].B_name = parentBytes
					block.B_content[indexContent].B_inodo = sb.S_inodes_count

					// Deserializar el bloque
					err = block.Serialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
					if err != nil {
						return false, err
					}

					content.B_name = parentBytes
					content.B_inodo = sb.S_inodes_count

					// Creamos el inodo newBloque
					inode := &Inode{
						I_uid:   1,
						I_gid:   1,
						I_size:  0,
						I_atime: float32(time.Now().Unix()),
						I_ctime: float32(time.Now().Unix()),
						I_mtime: float32(time.Now().Unix()),
						I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
						I_type:  [1]byte{'0'},
						I_perm:  [3]byte{'6', '6', '4'},
					}

					// Serializar el inodo users.txt
					err = inode.Serialize(path, int64(sb.S_first_ino))
					if err != nil {
						return false, err
					}

					// Actualizar el bitmap de inodos
					err = sb.UpdateBitmapInode(path)
					if err != nil {
						return false, err
					}

					sb.S_inodes_count++
					sb.S_free_inodes_count--
					sb.S_first_ino += sb.S_inode_size
				}

				// Obtenemos la carpeta padre más cercana
				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return false, err
				}

				// Convertir B_name a string y eliminar los caracteres nulos
				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				// Convertir parentDir a string y eliminar los caracteres nulos
				parentDirName := strings.Trim(parentDir, "\x00 ")
				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				if strings.EqualFold(contentName, parentDirName) {
					//fmt.Println("---------LA ENCONTRÉ-------")
					// Si son las mismas, entonces entramos al inodo que apunta el bloque
					creado, err := sb.createFolderInInode(path, p, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir)
					if err != nil {
						return false, err
					}
					return creado, nil
				}
			} else {
				//fmt.Println("---------ESTOY  CREANDO--------")

				// Si el apuntador al inodo está ocupado, continuar con el siguiente
				if content.B_inodo != -1 {
					continue
				}

				// Actualizar el contenido del bloque
				copy(content.B_name[:], destDir)
				content.B_inodo = sb.S_inodes_count

				// Actualizar el bloque
				block.B_content[indexContent] = content

				// Serializar el bloque
				err = block.Serialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return false, err
				}

				// Crear el inodo de la carpeta
				folderInode := &Inode{
					I_uid:   1,
					I_gid:   1,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				// Serializar el inodo de la carpeta
				err = folderInode.Serialize(path, int64(sb.S_first_ino))
				if err != nil {
					return false, err
				}

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInode(path)
				if err != nil {
					return false, err
				}

				// Actualizar el superbloque
				sb.S_inodes_count++
				sb.S_free_inodes_count--
				sb.S_first_ino += sb.S_inode_size

				// Crear el bloque de la carpeta
				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: content.B_inodo},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				// Serializar el bloque de la carpeta
				err = folderBlock.Serialize(path, int64(sb.S_first_blo))
				if err != nil {
					return false, err
				}

				// Actualizar el bitmap de bloques
				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return false, err
				}

				// Actualizar el superbloque
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size

				return true, nil
			}
		}

	}
	return false, nil
}

// createFolderInInode crea una carpeta en un inodo específico
func (sb *SuperBlock) createFileInInode(path string, r bool, inodeIndex int32, parentsDir []string, destDir string) (*Inode, bool, error) {
	// Crear un nuevo inodo
	inode := &Inode{}
	// Deserializar el inodo
	err := inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return inode, false, err
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] == '1' {
		return inode, false, nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for i, blockIndex := range inode.I_block {

		// Si el bloque no existe, salir
		if blockIndex == -1 {
			if len(parentsDir) == 0 {
				var arch [12]byte

				copy(arch[:], []byte(destDir))

				// Creamos el bloque del Inodo Raíz
				Block := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: 0},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				Block.B_content[2].B_name = arch
				Block.B_content[2].B_inodo = sb.S_inodes_count

				// Serializar el bloque de users.txt
				err = Block.Serialize(path, int64(sb.S_first_blo))
				if err != nil {
					return inode, false, err
				}

				//Actualizar el inodo
				inode.I_block[i] = sb.S_blocks_count
				err = inode.ActualizarI_mtime(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
				if err != nil {
					return inode, false, err
				}

				// Creamos el inodo newBloque
				inodeNew := &Inode{
					I_uid:   1,
					I_gid:   1,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'1'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				// Serializar
				err = inodeNew.Serialize(path, int64(sb.S_first_ino))
				if err != nil {
					return inode, false, err
				}

				//Actualizar todo lo referente al super bloque

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInode(path)
				if err != nil {
					return inode, false, err
				}

				// Actualizar el bitmap de bloques
				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return inode, false, err
				}

				// Actualizamos el superbloque
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size

				sb.S_inodes_count++
				sb.S_free_inodes_count--
				sb.S_first_ino += sb.S_inode_size

				return inodeNew, true, nil

			}
			if r {
				var carpeta [12]byte

				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return inode, false, err
				}
				copy(carpeta[:], []byte(parentDir))

				// Creamos el bloque del Inodo Raíz
				Block := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: 0},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				Block.B_content[2].B_name = carpeta
				Block.B_content[2].B_inodo = sb.S_inodes_count

				// Serializar el bloque de users.txt
				err = Block.Serialize(path, int64(sb.S_first_blo))
				if err != nil {
					return inode, false, err
				}

				//Actualizar el inodo
				inode.I_block[i] = sb.S_blocks_count
				err = inode.ActualizarI_mtime(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
				if err != nil {
					return inode, false, err
				}

				// Creamos el inodo newBloque
				inodeNew := &Inode{
					I_uid:   1,
					I_gid:   1,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				// Serializar
				err = inodeNew.Serialize(path, int64(sb.S_first_ino))
				if err != nil {
					return inode, false, err
				}

				// Agregarle un valor artificial a blockIndes
				blockIndex = sb.S_blocks_count

				//Actualizar todo lo referente al super bloque

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInode(path)
				if err != nil {
					return inode, false, err
				}

				// Actualizar el bitmap de bloques
				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return inode, false, err
				}

				// Actualizamos el superbloque
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size

				sb.S_inodes_count++
				sb.S_free_inodes_count--
				sb.S_first_ino += sb.S_inode_size

			} else {
				return inode, false, errors.New("No se encuentra en el directorio las carpetas padre y no se porpociono el parametro r")
			}

		}

		// Crear un nuevo bloque de carpeta
		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return inode, false, err
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.B_content[indexContent]

			// Sí las carpetas padre no están vacías debereamos buscar la carpeta padre más cercana
			if len(parentsDir) != 0 {
				//fmt.Println("---------ESTOY  VISITANDO--------")

				// Si el contenido está vacío, salir
				if content.B_inodo == -1 {
					if !r {
						return inode, false, errors.New("No se encuentra en el directorio las carpetas padre y no se porpociono el parametro r")
					}

					// Obtenemos la carpeta padre más cercana
					parentDir, err := utils.First(parentsDir)
					if err != nil {
						return inode, false, err
					}

					var parentBytes [12]byte
					copy(parentBytes[:], []byte(parentDir))

					block.B_content[indexContent].B_name = parentBytes
					block.B_content[indexContent].B_inodo = sb.S_inodes_count

					// Deserializar el bloque
					err = block.Serialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
					if err != nil {
						return inode, false, err
					}

					content.B_name = parentBytes
					content.B_inodo = sb.S_inodes_count

					// Creamos el inodo newBloque
					inode := &Inode{
						I_uid:   1,
						I_gid:   1,
						I_size:  0,
						I_atime: float32(time.Now().Unix()),
						I_ctime: float32(time.Now().Unix()),
						I_mtime: float32(time.Now().Unix()),
						I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
						I_type:  [1]byte{'0'},
						I_perm:  [3]byte{'6', '6', '4'},
					}

					// Serializar el inodo users.txt
					err = inode.Serialize(path, int64(sb.S_first_ino))
					if err != nil {
						return inode, false, err
					}

					// Actualizar el bitmap de inodos
					err = sb.UpdateBitmapInode(path)
					if err != nil {
						return inode, false, err
					}

					sb.S_inodes_count++
					sb.S_free_inodes_count--
					sb.S_first_ino += sb.S_inode_size
				}

				// Obtenemos la carpeta padre más cercana
				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return inode, false, err
				}

				// Convertir B_name a string y eliminar los caracteres nulos
				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				// Convertir parentDir a string y eliminar los caracteres nulos
				parentDirName := strings.Trim(parentDir, "\x00 ")
				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				if strings.EqualFold(contentName, parentDirName) {
					//fmt.Println("---------LA ENCONTRÉ-------")
					// Si son las mismas, entonces entramos al inodo que apunta el bloque
					creado, err := sb.createFolderInInode(path, r, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir)
					if err != nil {
						return inode, false, err
					}
					return inode, creado, nil
				}
			} else {
				//fmt.Println("---------ESTOY  CREANDO--------")

				// Si el apuntador al inodo está ocupado, continuar con el siguiente
				if content.B_inodo != -1 {
					continue
				}

				// Actualizar el contenido del bloque
				copy(content.B_name[:], destDir)
				content.B_inodo = sb.S_inodes_count

				// Actualizar el bloque
				block.B_content[indexContent] = content

				// Serializar el bloque
				err = block.Serialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return inode, false, err
				}

				// Crear el inodo de la carpeta
				folderInode := &Inode{
					I_uid:   1,
					I_gid:   1,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'1'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				// Serializar el inodo de la carpeta
				err = folderInode.Serialize(path, int64(sb.S_first_ino))
				if err != nil {
					return inode, false, err
				}

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInode(path)
				if err != nil {
					return inode, false, err
				}

				// Actualizar el superbloque
				sb.S_inodes_count++
				sb.S_free_inodes_count--
				sb.S_first_ino += sb.S_inode_size

				// // Crear el bloque de la carpeta
				// folderBlock := &FolderBlock{
				// 	B_content: [4]FolderContent{
				// 		{B_name: [12]byte{'.'}, B_inodo: content.B_inodo},
				// 		{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
				// 		{B_name: [12]byte{'-'}, B_inodo: -1},
				// 		{B_name: [12]byte{'-'}, B_inodo: -1},
				// 	},
				// }

				// // Serializar el bloque de la carpeta
				// err = folderBlock.Serialize(path, int64(sb.S_first_blo))
				// if err != nil {
				// 	return inode, false, err
				// }

				// // Actualizar el bitmap de bloques
				// err = sb.UpdateBitmapBlock(path)
				// if err != nil {
				// 	return inode, false, err
				// }

				// // Actualizar el superbloque
				// sb.S_blocks_count++
				// sb.S_free_blocks_count--
				// sb.S_first_blo += sb.S_block_size

				return inode, true, nil
			}
		}

	}
	return inode, false, nil
}
