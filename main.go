package main

// build requires libstoragemgmt-devel and libstoragemgmt installed (for cgo integration)
// execution requires libstoragemgmt installed on the host
//
// Build with the following command
// CGO_LDFLAGS=/usr/lib64/libstoragemgmt.so go build -o localdisk
//

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	lsm "github.com/libstorage/libstoragemgmt-golang"
	localdisk "github.com/libstorage/libstoragemgmt-golang/localdisk"
)

var healthText = map[lsm.DiskHealthStatus]string{
	lsm.DiskHealthStatusUnknown: "Unknown",
	lsm.DiskHealthStatusFail:    "Fail",
	lsm.DiskHealthStatusWarn:    "Warn",
	lsm.DiskHealthStatusGood:    "Good",
}
var linkText = map[lsm.DiskLinkType]string{
	lsm.DiskLinkTypeNoSupport: "Not supported by LSM",
	lsm.DiskLinkTypeUnknown:   "Unknown",
	lsm.DiskLinkTypeFc:        "FibreChannel",
	lsm.DiskLinkTypeSsa:       "SSA",
	lsm.DiskLinkTypeSbp:       "Serial Bus Protocol",
	lsm.DiskLinkTypeSrp:       "SCSI RDMA",
	lsm.DiskLinkTypeIscsi:     "iSCSI",
	lsm.DiskLinkTypeSas:       "SAS",
	lsm.DiskLinkTypeAdt:       "Automated Drive(Tape)",
	lsm.DiskLinkTypeAta:       "IDE/SATA",
	lsm.DiskLinkTypeUsb:       "USB",
	lsm.DiskLinkTypeSop:       "SCSI over PCIe",
	lsm.DiskLinkTypePciE:      "PCIe",
}

type disk struct {
	devPath      string
	devType      string
	serialNumber string
	vpd83        string
	sizeBytes    int64
	sizeSectors  int64
	transport    string
	linkSpeed    uint32
	rpm          int32
	ledIdent     string
	ledFail      string
	health       string
	model        string
	revision     string
	vendor       string
	wwid         string
	sectorFormat string
}

// bytesToHuman : convert a bytes value to a human readable format
func bytesToHuman(b int64) string {
	//
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// ConvertLedStatus : return a text value for an led state bit
func convertLedStatus(bitField lsm.DiskLedStatusBitField, offset uint) (string, error) {
	var ledText string
	ledState := bitField >> offset
	switch ledState {
	case 1:
		ledText = "ON"
	case 2:
		ledText = "OFF"
	case 4:
		ledText = "UNKNOWN"
	default:
		ledText = "UNKNOWN"
	}
	return ledText, nil
}

// readFile: read contents of a given filepath
func readFile(fname string) (string, error) {
	dat, err := ioutil.ReadFile(fname)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(dat)), nil
}

// extractDev : extract device name from a path
func extractDev(devicePath string) (string, error) {
	var components []string

	components = strings.Split(devicePath, "/")
	if len(components) < 2 {
		return "", errors.New("Invalid pathname of " + devicePath + "received")
	}

	return components[len(components)-1], nil
}

// getDeviceAttr : read sysfs for a device attribute
func getDeviceAttr(devicePath string, attr string) (string, error) {
	var err error

	devName, _ := extractDev(devicePath)
	sysfsPath := "/sys/class/block/" + devName + "/device/" + attr
	content, err := readFile(sysfsPath)
	if err != nil {
		return "", errors.New("attribute read error for " + attr + " on device " + devName)
	}
	return content, nil
}

// getDiskInfo : use lsm to get disk metadata
func getDiskInfo(devPath string, disk *disk) error {

	var health lsm.DiskHealthStatus
	var linkType lsm.DiskLinkType
	var ledStatus lsm.DiskLedStatusBitField

	if _, err := os.Stat(devPath); os.IsNotExist(err) {
		return errors.New("Device path not found")
	}

	disk.devPath = devPath
	disk.serialNumber, _ = localdisk.SerialNumGet(devPath)

	// We supplement the data available from LSM with direct queries into sysfs
	devName, _ := extractDev(devPath)
	sizeStr, _ := getDeviceAttr(devPath, "/block/"+devName+"/size")
	disk.sizeSectors, _ = strconv.ParseInt(sizeStr, 10, 64)
	disk.model, _ = getDeviceAttr(devPath, "model")
	disk.vendor, _ = getDeviceAttr(devPath, "vendor")
	disk.wwid, _ = getDeviceAttr(devPath, "wwid")
	disk.revision, _ = getDeviceAttr(devPath, "rev")
	physicalSector, _ := getDeviceAttr(devPath, "/block/"+devName+"/queue/physical_block_size")
	logicalSector, _ := getDeviceAttr(devPath, "/block/"+devName+"/queue/logical_block_size")
	if logicalSector == physicalSector {
		if logicalSector == "512" {
			disk.sectorFormat = "512"
		} else {
			disk.sectorFormat = "4KN"
		}
	} else {
		disk.sectorFormat = "512e"
	}
	if disk.sectorFormat == "4KN" {
		logicalSectorInt, _ := strconv.ParseInt(logicalSector, 10, 64)
		disk.sizeBytes = disk.sizeSectors * logicalSectorInt
	} else {
		disk.sizeBytes = disk.sizeSectors * 512
	}

	health, _ = localdisk.HealthStatusGet(devPath)
	disk.health = healthText[health]
	disk.rpm, _ = localdisk.RpmGet(devPath)

	switch disk.rpm {
	case 0:
		disk.devType = "Flash"
	default:
		disk.devType = "HDD"
	}

	disk.vpd83, _ = localdisk.Vpd83Get(devPath)
	disk.linkSpeed, _ = localdisk.LinkSpeedGet(devPath)
	linkType, _ = localdisk.LinkTypeGet(devPath)
	disk.transport = linkText[linkType]
	ledStatus, _ = localdisk.LedStatusGet(devPath)

	// Testing:
	// ledStatus = lsm.DiskLedStatusBitField(0x0000000000000004)
	if ledStatus == 1 {
		disk.ledIdent = "Unavailable"
		disk.ledFail = "Unavailable"
	} else {
		disk.ledIdent, _ = convertLedStatus(ledStatus, 1)
		disk.ledFail, _ = convertLedStatus(ledStatus, 4)
	}

	return nil
}

// setFailLed : set/unset the fail led
func setFailLed(devPath string, state string) error {

	switch state {
	case "on":
		return localdisk.FaultLedOn(devPath)
	case "off":
		return localdisk.FaultLedOff(devPath)
	default:
		return nil
	}
}

// listDisks : show all the local disks on the system
func listDisks() {
	var disks []string
	var disk disk

	disks, _ = localdisk.List()
	// TODO check if list has entries
	fmt.Println(fmt.Sprintf("%-16s %6s %-15s %15s %6s %10s %5s %9s %11s %11s %7s %16s %16s %8s %20s",
		"Device Path",
		"Type",
		"Serial Number",
		"Size",
		"Sector",
		"Transport",
		"RPM",
		"Bus Speed",
		"IDENT",
		"FAIL",
		"Health",
		"Vendor",
		"Model",
		"Revision",
		"wwid"))

	for _, devPath := range disks {
		_ = getDiskInfo(devPath, &disk)

		fmt.Println(fmt.Sprintf("%-16s %6s %-15s %15s %6s %10s %5d %9d %11s %11s %7s %16s %16s %8s %20s",
			disk.devPath,
			disk.devType,
			disk.serialNumber,
			bytesToHuman(disk.sizeBytes),
			disk.sectorFormat,
			disk.transport,
			disk.rpm,
			disk.linkSpeed,
			disk.ledIdent,
			disk.ledFail,
			disk.health,
			disk.vendor,
			disk.model,
			disk.revision,
			disk.wwid))
	}
}

// showDisk : Show details for a specific disk
func showDisk(devPath string) {
	var disk disk
	var err error

	err = getDiskInfo(devPath, &disk)
	if err != nil {
		fmt.Println("Unable to list device " + devPath)
		os.Exit(1)
	}

	fmt.Printf("Device Path    : %s\n", (disk.devPath))
	fmt.Printf("Type           : %s\n", (disk.devType))
	fmt.Printf("Serial Number  : %s\n", (disk.serialNumber))
	fmt.Printf("Size           : %s\n", (bytesToHuman(disk.sizeBytes)))
	fmt.Printf("Sector Format  : %s\n", (disk.sectorFormat))
	fmt.Printf("Transport      : %s\n", (disk.transport))
	fmt.Printf("RPM            : %d\n", (disk.rpm))
	fmt.Printf("Bus Speed      : %d\n", (disk.linkSpeed))
	fmt.Printf("IDENT LED      : %s\n", (disk.ledIdent))
	fmt.Printf("FAIL LED       : %s\n", (disk.ledFail))
	fmt.Printf("Health         : %s\n", (disk.health))
	fmt.Printf("Vendor         : %s\n", (disk.vendor))
	fmt.Printf("Model          : %s\n", (disk.model))
	fmt.Printf("Revision       : %s\n", (disk.revision))
	fmt.Printf("wwid           : %s\n", (disk.wwid))
}

func main() {
	var err error

	listDisksPtr := flag.Bool("list", false, "list all local disks")
	getDiskPtr := flag.String("show", "", "show a specific disk matching given /dev name")
	setFailOnPtr := flag.String("fail-led-on", "", "activate fail LED on a given device")
	setFailOffPtr := flag.String("fail-led-off", "", "de-activate fail LED on a given device")
	flag.Parse()

	if *listDisksPtr {
		listDisks()
		os.Exit(0)
	}
	if *getDiskPtr != "" {
		showDisk(*getDiskPtr)
		os.Exit(0)
	}
	if *setFailOnPtr != "" || *setFailOffPtr != "" {
		if *setFailOnPtr != "" {
			err = setFailLed(*setFailOnPtr, "on")
		}
		if *setFailOffPtr != "" {
			err = setFailLed(*setFailOffPtr, "off")
		}
		if err != nil {
			fmt.Println("Unable to set the disks fault LED beacon")
		}
	}
}
