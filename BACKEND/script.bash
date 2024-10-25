rmdisk -path=/home/fernando/Escritorio/Reportes/disks/Disco.mia
mkdisk -size=5 -unit=M -fit=WF -path=/home/fernando/Escritorio/Reportes/disks/Disco.mia
fdisk -size=1 -type=P -unit=K -fit=BF -name="Particion1" -path="/home/fernando/Escritorio/Reportes/disks/Disco.mia"
fdisk -size=1 -type=E -unit=M -fit=BF -name="Particion2" -path="/home/fernando/Escritorio/Reportes/disks/Disco.mia"
fdisk -size=1 -type=P -unit=K -fit=BF -name="Particion3" -path="/home/fernando/Escritorio/Reportes/disks/Disco.mia"
fdisk -size=1 -type=P -unit=M -fit=BF -name="Particion4" -path="/home/fernando/Escritorio/Reportes/disks/Disco.mia"
fdisk -size=1 -type=L -unit=K -fit=BF -name="Particion5" -path="/home/fernando/Escritorio/Reportes/disks/Disco.mia"
fdisk -size=1 -type=L -unit=K -fit=BF -name="Particion6" -path="/home/fernando/Escritorio/Reportes/disks/Disco.mia"
fdisk -size=500 -type=L -unit=K -fit=BF -name="Particion7" -path="/home/fernando/Escritorio/Reportes/disks/Disco.mia"
mount -path="/home/fernando/Escritorio/Reportes/disks/Disco.mia" -name=Particion1
mount -path="/home/fernando/Escritorio/Reportes/disks/Disco.mia" -name=Particion3
mount -path="/home/fernando/Escritorio/Reportes/disks/Disco.mia" -name=Particion4
mkfs -id=153A -type=full
rep -id=151A -path="/home/fernando/Escritorio/Reportes/RepMBE.jpg"  -name=mbr
rep -id=151A -path="/home/fernando/Escritorio/Reportes/RepDisk.jpg"  -name=disk
login -user=root -pass=123 -id=153A
mkgrp -name=usuarios
mkgrp -name=usuarios1
mkgrp -name=usuarios2
mkgrp -name=usuarios3
mkgrp -name=usuarios4
mkgrp -name=usuarios5
mkgrp -name=usuarios6
mkgrp -name=usuarios8
mkgrp -name=usuarios9
mkgrp -name=usuarios10
mkusr -user=user1 -pass=6965 -grp=usuarios3
mkusr -user=user2 -pass=1236589 -grp=usuarios45
rep -id=153A -path="/home/fernando/Escritorio/Reportes/inode.jpg"  -name=inode
rep -id=153A -path="/home/fernando/Escritorio/Reportes/bmInode.txt "  -name=bm_inode