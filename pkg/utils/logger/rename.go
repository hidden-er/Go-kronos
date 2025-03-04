package logger

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"Chamael/internal/party"
	"Chamael/pkg/config"
)

func RenameCommon(c config.CommonConfig, p party.CommonParty, path string) {
	dir_send := fmt.Sprintf("%s%s", path, c.IPList[p.PID]+":"+c.PortList[p.PID])
	newdir_send := fmt.Sprintf("%snode%d", path, p.PID)
	os.Rename(dir_send, newdir_send)

	file_recv := fmt.Sprintf("%s(Received)0.0.0.0:%s.log", path, c.PortList[p.PID])
	newfile_recv := fmt.Sprintf("%s(Received)node%d", path, p.PID)
	os.Rename(file_recv, newfile_recv)

	files, _ := ioutil.ReadDir(newdir_send)
	for _, file := range files {
		oldName := file.Name()
		trimmed := strings.TrimPrefix(oldName, "(Send)")
		trimmed = strings.TrimSuffix(trimmed, ".log")
		ipAndPort := strings.Split(trimmed, ":")
		if len(ipAndPort) != 2 {
			log.Printf("Skipping invalid file name: %s", oldName)
			continue
		}
		ip := ipAndPort[0]
		port := ipAndPort[1]
		var nodeNumber int
		for i := 0; i < c.N*c.M; i++ {
			if ip == c.IPList[i] && port == c.PortList[i] {
				nodeNumber = i
				break
			}
		}
		newName := fmt.Sprintf("(Sendto)node%d.log", nodeNumber)
		os.Rename(newdir_send+"/"+oldName, newdir_send+"/"+newName)
	}
}

func RenameHonest(c config.HonestConfig, p party.HonestParty, path string) {
	dir_send := fmt.Sprintf("%s%s", path, c.IPList[p.PID]+":"+c.PortList[p.PID])
	newdir_send := fmt.Sprintf("%snode%d", path, p.PID)
	os.Rename(dir_send, newdir_send)

	file_recv := fmt.Sprintf("%s(Received)0.0.0.0:%s.log", path, c.PortList[p.PID])
	newfile_recv := fmt.Sprintf("%s(Received)node%d", path, p.PID)
	os.Rename(file_recv, newfile_recv)

	files, _ := ioutil.ReadDir(newdir_send)
	for _, file := range files {
		oldName := file.Name()
		trimmed := strings.TrimPrefix(oldName, "(Send)")
		trimmed = strings.TrimSuffix(trimmed, ".log")
		ipAndPort := strings.Split(trimmed, ":")
		if len(ipAndPort) != 2 {
			log.Printf("Skipping invalid file name: %s", oldName)
			continue
		}
		ip := ipAndPort[0]
		port := ipAndPort[1]
		var nodeNumber int
		for i := 0; i < c.N*c.M; i++ {
			if ip == c.IPList[i] && port == c.PortList[i] {
				nodeNumber = i
				break
			}
		}
		newName := fmt.Sprintf("(Sendto)node%d.log", nodeNumber)
		os.Rename(newdir_send+"/"+oldName, newdir_send+"/"+newName)
	}
}
