package main

// build requires libstoragemgmt-devel and libstoragemgmt installed (for cgo integration)
// execution requires libstoragemgmt installed on the host
//
// Build with the following command
// CGO_LDFLAGS=/usr/lib64/libstoragemgmt.so go build -o localdisk
//

import (
	"flag"
	"fmt"
	"io/ioutil"
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
	serialNumber string
	vpd83        string
	size         int64
	linkType     string
	linkSpeed    uint32
	rpm          int32
	ledIdent     string
	ledFail      string
	health       string
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

// DiskSize : extract disk size (blocks) from sysfs path
func diskSize(devicePath string) (int64, error) {
	var components []string
	var path string
	var content string
	var size int64

	components = strings.Split(devicePath, "/")
	devName := components[len(components)-1]

	path = "/sys/class/block/" + devName + "/size"
	content, _ = readFile(path)
	size, _ = strconv.ParseInt(content, 10, 64)
	return size, nil
}

// getDiskInfo : use lsm to get disk metadata
func getDiskInfo(devPath string, disk *disk) error {

	var health lsm.DiskHealthStatus
	var linkType lsm.DiskLinkType
	var ledStatus lsm.DiskLedStatusBitField

	disk.devPath = devPath
	disk.serialNumber, _ = localdisk.SerialNumGet(devPath)
	disk.size, _ = diskSize(devPath)
	health, _ = localdisk.HealthStatusGet(devPath) // -1 = unknown, 2 Good
	disk.health = healthText[health]
	disk.rpm, _ = localdisk.RpmGet(devPath)
	disk.vpd83, _ = localdisk.Vpd83Get(devPath)
	disk.linkSpeed, _ = localdisk.LinkSpeedGet(devPath)
	linkType, _ = localdisk.LinkTypeGet(devPath)
	disk.linkType = linkText[linkType]
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

// ShowDisks : show all the local disks on the system
func showDisks() {
	var disks []string
	var disk disk

	disks, _ = localdisk.List()
	// TODO check if list has entries
	fmt.Println(fmt.Sprintf("%-16s %-15s %11s %8s %5s %9s %11s %11s %7s",
		"Device Path",
		"Serial Number",
		"Size",
		"Type",
		"RPM",
		"Bus Speed",
		"IDENT",
		"FAIL",
		"Health"))

	for _, devPath := range disks {
		_ = getDiskInfo(devPath, &disk)

		fmt.Println(fmt.Sprintf("%-16s %-15s %11d %8s %5d %9d %11s %11s %7s",
			disk.devPath,
			disk.serialNumber,
			disk.size,
			disk.linkType,
			disk.rpm,
			disk.linkSpeed,
			disk.ledIdent,
			disk.ledFail,
			disk.health))
	}
}

func main() {
	var err error

	listDisksPtr := flag.Bool("list", false, "list all local disks")
	setFailOnPtr := flag.String("disk-fail-led-on", "", "activate fail LED on a given device")
	setFailOffPtr := flag.String("disk-fail-led-off", "", "de-activate fail LED on a given device")
	flag.Parse()

	if *listDisksPtr {
		showDisks()
	} else {
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
