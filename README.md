# localdisk
Example project showing how the libstoragemgmt-golang package can be used to interact with the libstoragemgmt library. The libstoragemgmt library (LSM), provides interfaces to perform common storage tasks across a variety of backends. This example code, focuses solely on localdisk interaction.

Obviously, as an admin you can just run the lsmcli commands directly, but this project shows how you can interact with libstoragemgmt programmatically for automation, or platform integration.

## Prerequisites
1. For build and testing you need the following rpms installed on your system
- libstoragemgmt
- libstoragemgmt-devel
2. For running on a host the host must have the libstoragemgmt rpm installed.  

3. To run either lsmcli or this code, the user requires r/w privileges to the device (root normally)  

## Build
The localdisk package within libstoragemgmt-golang uses **cgo** to interact with the libstoragemgmt api, so to build you need to point the linker at the local machines install location of the libstoragemgmt library  
e.g.

```
# CGO_LDFLAGS=/usr/lib64/libstoragemgmt.so go build -o localdisk
```

## Running the tool
1. Example options
```
[root@srv-01 bin]# localdisk -h
Usage of localdisk:
  -fail-led-off string
    	de-activate fail LED on a given device
  -fail-led-on string
    	activate fail LED on a given device
  -list
    	list all local disks
  -show string
    	show a specific disk matching given /dev name

```
2. Turn on fail LED
```
localdisk -fail-led-on /dev/sda
```
3. Turn off fail LED
```
localdisk -fail-led-off /dev/sda
```

## Output Examples
1. Disk list
```
[root@srv-01 ~]# localdisk -list
Device Path        Type Serial Number              Size Sector  Transport   RPM Bus Speed       IDENT        FAIL  Health           Vendor            Model Revision                 wwid
/dev/sda            HDD 15P0A0R0FRD6          279.4 GiB    512        SAS 10000      6000     UNKNOWN         OFF    Good          TOSHIBA       AL13SEB300     DE0D naa.50000396082ba631
/dev/sdb            HDD 15P0A0YFFRD6          279.4 GiB    512        SAS 10000      6000     UNKNOWN         OFF    Good          TOSHIBA       AL13SEB300     DE0D naa.50000396082bbbf9
/dev/sdk            HDD 15P0A0ONFRD6          279.4 GiB    512        SAS 10000      6000     UNKNOWN         OFF    Good          TOSHIBA       AL13SEB300     DE0D naa.50000396082b989d
/dev/sdl            HDD 15P0A0YBFRD6          279.4 GiB    512        SAS 10000      6000     UNKNOWN         OFF    Good          TOSHIBA       AL13SEB300     DE0D naa.50000396082bb9d1
/dev/sdm          Flash BTWL452503K7480QGN       447.1 GiB   512e   IDE/SATA     0      6000     UNKNOWN         OFF Unknown              ATA INTEL SSDSC2BB48     DL13 naa.55cd2e404b753fb0
/dev/sdn          Flash BTWL452503PJ480QGN       447.1 GiB   512e   IDE/SATA     0      6000     UNKNOWN         OFF Unknown              ATA INTEL SSDSC2BB48     DL13 naa.55cd2e404b754043
/dev/sdo          Flash BTWL452503K2480QGN       447.1 GiB   512e   IDE/SATA     0      6000     UNKNOWN         OFF Unknown              ATA INTEL SSDSC2BB48     DL13 naa.55cd2e404b753fab
/dev/sdp          Flash BTWL452503PF480QGN       447.1 GiB   512e   IDE/SATA     0      6000     UNKNOWN         OFF Unknown              ATA INTEL SSDSC2BB48     DL13 naa.55cd2e404b754040
/dev/sdc            HDD 15R0A08WFRD6          279.4 GiB    512        SAS 10000      6000     UNKNOWN         OFF    Good          TOSHIBA       AL13SEB300     DE0D naa.500003960831c74d
/dev/sdd            HDD 15R0A07DFRD6          279.4 GiB    512        SAS 10000      6000     UNKNOWN         OFF    Good          TOSHIBA       AL13SEB300     DE0D naa.500003960831bfbd
/dev/sde            HDD 15P0A0QDFRD6          279.4 GiB    512        SAS 10000      6000     UNKNOWN         OFF    Good          TOSHIBA       AL13SEB300     DE0D naa.50000396082ba3a1
/dev/sdf            HDD 15R0A064FRD6          279.4 GiB    512        SAS 10000      6000     UNKNOWN         OFF    Good          TOSHIBA       AL13SEB300     DE0D naa.500003960831b065
/dev/sdg            HDD 15P0A0QWFRD6          279.4 GiB    512        SAS 10000      6000     UNKNOWN         OFF    Good          TOSHIBA       AL13SEB300     DE0D naa.50000396082ba5fd
/dev/sdh            HDD 15P0A0O8FRD6          279.4 GiB    512        SAS 10000      6000     UNKNOWN         OFF    Good          TOSHIBA       AL13SEB300     DE0D naa.50000396082b9675
/dev/sdi            HDD 15P0A0RFFRD6          279.4 GiB    512        SAS 10000      6000     UNKNOWN         OFF    Good          TOSHIBA       AL13SEB300     DE0D naa.50000396082ba709
/dev/sdj            HDD 15R0A07PFRD6          279.4 GiB    512        SAS 10000      6000     UNKNOWN         OFF    Good          TOSHIBA       AL13SEB300     DE0D naa.500003960831c051

```
2. Turning the fail LED ON
```
[root@srv-01 bin]# localdisk -fail-led-on /dev/sdf
[root@srv-01 bin]# localdisk -show /dev/sdf
Device Path    : /dev/sdf
Type           : HDD
Serial Number  : 15R0A064FRD6
Size           : 279.4 GiB
Sector Format  : 512
Transport      : SAS
RPM            : 10000
Bus Speed      : 6000
IDENT LED      : UNKNOWN
FAIL LED       : ON                   <-----
Health         : Good
Vendor         : TOSHIBA
Model          : AL13SEB300
Revision       : DE0D
wwid           : naa.500003960831b065
```
3. Turning the fail LED OFF
```
[root@srv-01 ~]# localdisk -fail-led-off /dev/sdf
[root@srv-01 ~]# localdisk -show /dev/sdf
Device Path    : /dev/sdf
Type           : HDD
Serial Number  : 15R0A064FRD6
Size           : 279.4 GiB
Sector Format  : 512
Transport      : SAS
RPM            : 10000
Bus Speed      : 6000
IDENT LED      : UNKNOWN
FAIL LED       : OFF                  <-----
Health         : Good
Vendor         : TOSHIBA
Model          : AL13SEB300
Revision       : DE0D
wwid           : naa.500003960831b065

```
  
After running this process, the changes to the fault LED could be seen in the server's BMC  
  
![LED-Changes](images/fault-led-test.png)

The CLI and GUI output shown above is from an old Dell r730 server, running RHEL7.4. It's reasonable to expect a more modern server to provide better information and support for features like disk IDENT and flash drive health.