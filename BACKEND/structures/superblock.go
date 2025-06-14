package structures

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type SuperBlock struct {
	S_filesystem_type   int32
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_inodes_count int32
	S_free_blocks_count int32
	S_mtime             float32
	S_umtime            float32
	S_mnt_count         int32
	S_magic             int32
	S_inode_size        int32
	S_block_size        int32
	S_first_ino         int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32
	// Total: 68 bytes
}

// Serialize escribe la estructura SuperBlock en un archivo binario en la posición especificada
func (sb *SuperBlock) Serialize(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Serializar la estructura SuperBlock directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, sb)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize lee la estructura SuperBlock desde un archivo binario en la posición especificada
func (sb *SuperBlock) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Obtener el tamaño de la estructura SuperBlock
	sbSize := binary.Size(sb)
	if sbSize <= 0 {
		return fmt.Errorf("invalid SuperBlock size: %d", sbSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura SuperBlock
	buffer := make([]byte, sbSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura SuperBlock
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, sb)
	if err != nil {
		return err
	}

	return nil
}

// Crear users.txt
func (sb *SuperBlock) CrearUsersFile(path string) error {
	// ----------- Creamos / -----------
	// Creamos el inodo raíz
	rootInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  0,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Serializar el inodo raíz
	err := rootInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de inodos
	err = sb.UpdateBitmapInode(path)
	if err != nil {
		return err
	}

	// Actualizar el superbloque
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// Creamos el bloque del Inodo Raíz
	rootBlock := &FolderBlock{
		B_content: [4]FolderContent{
			{B_name: [12]byte{'.'}, B_inodo: 0},
			{B_name: [12]byte{'.', '.'}, B_inodo: 0},
			{B_name: [12]byte{'-'}, B_inodo: -1},
			{B_name: [12]byte{'-'}, B_inodo: -1},
		},
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.S_first_blo))
	if err != nil {
		return err
	}

	// Actualizar el superbloque
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	// // Verificar el inodo raíz
	// fmt.Println("\nInodo Raíz:")
	// rootInode.Print()

	// // Verificar el bloque de carpeta raíz
	// fmt.Println("\nBloque de Carpeta Raíz:")
	// rootBlock.Print()

	// ----------- Creamos /users.txt -----------
	usersText := "1,G,root\n1,U,root,root,123\n"

	// Deserializar el inodo raíz
	err = rootInode.Deserialize(path, int64(sb.S_inode_start+0)) // 0 porque es el inodo raíz
	if err != nil {
		return err
	}

	// Actualizamos el inodo raíz
	rootInode.I_atime = float32(time.Now().Unix())

	// Serializar el inodo raíz
	err = rootInode.Serialize(path, int64(sb.S_inode_start+0)) // 0 porque es el inodo raíz
	if err != nil {
		return err
	}

	// Deserializar el bloque de carpeta raíz
	err = rootBlock.Deserialize(path, int64(sb.S_block_start+0)) // 0 porque es el bloque de carpeta raíz
	if err != nil {
		return err
	}

	// Actualizamos el bloque de carpeta raíz
	rootBlock.B_content[2] = FolderContent{B_name: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, B_inodo: sb.S_inodes_count}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.S_block_start+0)) // 0 porque es el bloque de carpeta raíz
	if err != nil {
		return err
	}

	// Creamos el inodo users.txt
	usersInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(len(usersText)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'1'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Actualizar el bitmap de inodos
	err = sb.UpdateBitmapInode(path)
	if err != nil {
		return err
	}

	// Serializar el inodo users.txt
	err = usersInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil {
		return err
	}

	// Actualizamos el superbloque
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// Creamos el bloque de users.txt
	usersBlock := &FileBlock{
		B_content: [64]byte{},
	}
	// Copiamos el texto de usuarios en el bloque
	copy(usersBlock.B_content[:], usersText)

	// Serializar el bloque de users.txt
	err = usersBlock.Serialize(path, int64(sb.S_first_blo))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	// Actualizamos el superbloque
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	// // Verificar el inodo raíz
	// fmt.Println("\nInodo Raíz Actualizado:")
	// rootInode.Print()

	// // Verificar el bloque de carpeta raíz
	// fmt.Println("\nBloque de Carpeta Raíz Actualizado:")
	// rootBlock.Print()

	// // Verificar el inodo users.txt
	// fmt.Println("\nInodo users.txt:")
	// usersInode.Print()

	// // Verificar el bloque de users.txt
	// fmt.Println("\nBloque de users.txt:")
	// usersBlock.Print()

	return nil
}

// PrintSuperBlock imprime los valores de la estructura SuperBlock
func (sb *SuperBlock) Print() {
	// Convertir el tiempo de montaje a una fecha
	mountTime := time.Unix(int64(sb.S_mtime), 0)
	// Convertir el tiempo de desmontaje a una fecha
	unmountTime := time.Unix(int64(sb.S_umtime), 0)

	fmt.Printf("Filesystem Type: %d\n", sb.S_filesystem_type)
	fmt.Printf("Inodes Count: %d\n", sb.S_inodes_count)
	fmt.Printf("Blocks Count: %d\n", sb.S_blocks_count)
	fmt.Printf("Free Inodes Count: %d\n", sb.S_free_inodes_count)
	fmt.Printf("Free Blocks Count: %d\n", sb.S_free_blocks_count)
	fmt.Printf("Mount Time: %s\n", mountTime.Format(time.RFC3339))
	fmt.Printf("Unmount Time: %s\n", unmountTime.Format(time.RFC3339))
	fmt.Printf("Mount Count: %d\n", sb.S_mnt_count)
	fmt.Printf("Magic: %d\n", sb.S_magic)
	fmt.Printf("Inode Size: %d\n", sb.S_inode_size)
	fmt.Printf("Block Size: %d\n", sb.S_block_size)
	fmt.Printf("First Inode: %d\n", sb.S_first_ino)
	fmt.Printf("First Block: %d\n", sb.S_first_blo)
	fmt.Printf("Bitmap Inode Start: %d\n", sb.S_bm_inode_start)
	fmt.Printf("Bitmap Block Start: %d\n", sb.S_bm_block_start)
	fmt.Printf("Inode Start: %d\n", sb.S_inode_start)
	fmt.Printf("Block Start: %d\n", sb.S_block_start)
}

// Imprimir inodos
func (sb *SuperBlock) PrintInodes(path string) error {
	// Imprimir inodos
	fmt.Println("\nInodos\n----------------")
	// Iterar sobre cada inodo
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &Inode{}
		// Deserializar el inodo
		err := inode.Deserialize(path, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return err
		}
		// Imprimir el inodo
		fmt.Printf("\nInodo %d:\n", i)
		inode.Print()
	}

	return nil
}

// Impriir bloques
func (sb *SuperBlock) PrintBlocks(path string) error {
	// Imprimir bloques
	fmt.Println("\nBloques\n----------------")
	// Iterar sobre cada inodo
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &Inode{}
		// Deserializar el inodo
		err := inode.Deserialize(path, int64(sb.S_inode_start+(i*sb.S_inode_size)))
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
				block := &FolderBlock{}
				// Deserializar el bloque
				err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				// Imprimir el bloque
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
				continue

				// Si el inodo es de tipo archivo
			} else if inode.I_type[0] == '1' {
				block := &FileBlock{}
				// Deserializar el bloque
				err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				// Imprimir el bloque
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
				continue
			}

		}
	}

	return nil
}

// Fucion para validar las credenciales del usuario que se quiere logear
func (sb *SuperBlock) ValidateCredentials(name string, pass string, pathDisk string) (bool, USER, error) {

	_, lisUser, err := sb.OpGruopUserTXT(pathDisk)
	if err != nil {
		return false, lisUser[0], err
	}

	for _, user := range lisUser {
		if strings.EqualFold(user.Usuario, name) && strings.EqualFold(user.Contrasena, pass) {
			return true, user, nil
		}
	}

	return false, lisUser[0], nil
}

// Fucion para validar si existe un nombre de un grupo en especifico
func (sb *SuperBlock) ValidateGroup(name_group string, pathDisk string) (bool, int, int, error) {
	var rootInode Inode
	var inode Inode

	err := rootInode.Deserialize(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*0)))
	if err != nil {
		return false, -1, -1, err
	}

	err = inode.Deserialize(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
	if err != nil {
		return false, -1, -1, err
	}

	// Nos ayudara a evaluar si no se ha inicializado bien la particion o capaz dicha particion no contiene la configuracion EXT2
	if inode.I_type[0] != '1' {
		return false, -1, -1, errors.New("Ocurrio un error en los inodos")
	}

	listGroup, _, err := sb.splitGroupsUsers(inode, pathDisk)
	if err != nil {
		return false, -1, -1, err
	}

	for _, group := range listGroup {
		if strings.EqualFold(group.Grupo, name_group) {
			rootInode.ActualizarI_atime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*0)))
			inode.ActualizarI_atime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
			return true, -1, group.Pos, nil
		}
	}

	return false, len(listGroup) + 1, -1, nil
}

func (sb *SuperBlock) AddTXT(grupo string, pathDisk string) error {
	rootInode := &Inode{}
	inode := &Inode{}

	err := rootInode.Deserialize(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*0)))
	if err != nil {
		return err
	}

	err2 := inode.Deserialize(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
	if err2 != nil {
		return err2
	}

	// Nos ayudara a evaluar si no se ha inicializado bien la particion o capaz dicha particion no contiene la configuracion EXT2
	if inode.I_type[0] != '1' {
		return errors.New("Ocurrio un error en los inodos")
	}

	err3 := sb.inserInfo(inode, pathDisk, grupo)
	if err3 != nil {
		return err3
	}

	rootInode.ActualizarI_atime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*0)))

	return nil
}

// Fucion para validar si existe un nombre de un usuario en especifico
func (sb *SuperBlock) ValidateUser(name_user string, name_group string, pathDisk string) (int, error) {

	var rootInode Inode
	var inode Inode

	err := rootInode.Deserialize(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*0)))
	if err != nil {
		return -1, err
	}

	err = inode.Deserialize(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
	if err != nil {
		return -1, err
	}

	// Nos ayudara a evaluar si no se ha inicializado bien la particion o capaz dicha particion no contiene la configuracion EXT2
	if inode.I_type[0] != '1' {
		return -1, errors.New("Ocurrio un error en los inodos")
	}

	listGroup, listUser, err := sb.splitGroupsUsers(inode, pathDisk)
	if err != nil {
		return -1, err
	}

	for _, user := range listUser {
		if strings.EqualFold(user.Usuario, name_user) {
			rootInode.ActualizarI_atime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*0)))
			inode.ActualizarI_atime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
			return -1, errors.New("El nombre del usuario ya existe en el disco")
		}
	}

	for _, group := range listGroup {
		if strings.EqualFold(group.Grupo, name_group) {
			return len(listUser) + 1, nil
		}
	}

	rootInode.ActualizarI_atime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*0)))
	inode.ActualizarI_atime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))

	return -1, errors.New("No existe el grupo a el que se le quiere agregar a el usuario")
}

// Esta funcion sera la encargada deserializar los inodos corepondientes a /users.txt para optener la lista usuarios y grupos
func (sb *SuperBlock) OpGruopUserTXT(pathDisk string) ([]GROUP, []USER, error) {
	var rootInode Inode
	var inode Inode

	err := rootInode.Deserialize(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*0)))
	if err != nil {
		return nil, nil, err
	}

	err2 := inode.Deserialize(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
	if err2 != nil {
		return nil, nil, err2
	}

	// Nos ayudara a evaluar si no se ha inicializado bien la particion o capaz dicha particion no contiene la configuracion EXT2
	if inode.I_type[0] != '1' {
		return nil, nil, errors.New("Ocurrio un error en los inodos")
	}

	listGroup, lisUser, err := sb.splitGroupsUsers(inode, pathDisk)

	if err != nil {
		return nil, nil, err
	}

	err3 := rootInode.ActualizarI_atime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*0)))
	if err3 != nil {
		return nil, nil, err3
	}

	err4 := inode.ActualizarI_atime(pathDisk, int64(sb.S_inode_start+(sb.S_inode_size*1)))
	if err4 != nil {
		return nil, nil, err4
	}

	return listGroup, lisUser, nil
}

/// Area de crear crapetas y textos

// CreateFolder crea una carpeta en el sistema de archivos
func (sb *SuperBlock) CreateFolder(path string, p bool, parentsDir []string, destDir string) error {
	// Si parentsDir está vacío, solo trabajar con el primer inodo que sería el raíz "/"
	if len(parentsDir) == 0 {
		guardado, err := sb.createFolderInInode(path, p, 0, parentsDir, destDir)
		if err != nil {
			return err
		}
		if guardado {
			print("se esta guardadnod ")
			return nil
		}
	}

	// Iterar sobre cada inodo ya que se necesita buscar el inodo padre
	for i := int32(0); i < sb.S_inodes_count; i++ {
		guardado, err := sb.createFolderInInode(path, p, i, parentsDir, destDir)
		if err != nil {
			return err
		}
		if guardado {
			return nil
		}
	}

	return nil
}

// CreateFile crea un archivo en el sistema de archivos
func (sb *SuperBlock) CreateFile(pathDisk string, r bool, parentsDir []string, destFile string, cont string) error {

	// Si parentsDir está vacío, solo trabajar con el primer inodo que sería el raíz "/"
	if len(parentsDir) == 0 {
		idone, creada, err := sb.createFileInInode(pathDisk, r, 0, parentsDir, destFile)
		if err != nil {
			return err
		}
		if creada {
			return sb.inserInfo(idone, pathDisk, cont)
		}
	}

	// Iterar sobre cada inodo ya que se necesita buscar el inodo padre
	for i := int32(0); i < sb.S_inodes_count; i++ {
		idone, creada, err := sb.createFileInInode(pathDisk, r, i, parentsDir, destFile)
		if err != nil {
			return err
		}
		if creada {
			return sb.inserInfo(idone, pathDisk, cont)
		}
	}

	return nil
}
