// Beisen.ReceiveMail project main.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"net/mail"
	"os"
	"strconv"
	"strings"

	"github.com/jhillyerd/go.enmime"
	"github.com/taknb2nch/go-pop3"
)

const (
	DumpPath = "e:\\emails\\"

	AttPath = "e:\\emails\\atts\\"
)

func main() {
	var ip = ""
	var port = ""
	var user = ""
	var pass = ""

	fmt.Println("*********欢迎使用邮件备份工具v0.1beta*********")
	running := true
	reader := bufio.NewReader(os.Stdin)
	for running {
		fmt.Println("请输入用户名：")
		data, _, _ := reader.ReadLine()
		user = string(data)

		fmt.Println("请输入密码：")
		data, _, _ = reader.ReadLine()
		pass = string(data)

		fmt.Println("请输入pop3服务器地址：")
		data, _, _ = reader.ReadLine()
		ip = string(data)

		fmt.Println("请输入端口：")
		data, _, _ = reader.ReadLine()
		port = string(data)

		address := ip + ":" + port
		
		client, err := pop3.Dial(address)
		defer func() {
			client.Quit()
			client.Close()
		}()

		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		if err = client.User(user); err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		if err = client.Pass(pass); err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		var count int
		var size uint64

		if count, size, err = client.Stat(); err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		log.Printf("Count: %d, Size: %d\n", count, size)

		running = false
	}

	fmt.Println("确定要备份吗?备份文件会存放在e:\\emails文件夹。(y/n)")
	data, _, _ := reader.ReadLine()
	commond := string(data)

	if !(commond == "y" || commond == "Y") {
		return
	}

	if !Exist(DumpPath) {
		os.Mkdir(DumpPath, os.ModeDir)
	}
	if !Exist(AttPath) {
		os.Mkdir(AttPath, os.ModeDir)
	}
	address := ip + ":" + port
	if err := pop3.ReceiveMail(address, user, pass,
		func(number int, uid, data string, err error) (bool, error) {
			//log.Printf("%d, %s\n", number, uid)
			numberStr := strconv.Itoa(number)
			file := strings.NewReader(data)
			msg, err := mail.ReadMessage(file)
			if err != nil {
				log.Println("解析失败", err, number, uid)
				return false, nil
			}
			mime, err := enmime.ParseMIMEBody(msg)
			if err != nil {
				log.Println("解析失败", err, number, uid)
				return false, nil
			}

			var subject = mime.GetHeader("Subject")
			log.Printf("下载：" + subject)
			userFile := DumpPath + numberStr + "," + subject + "(" + uid + ").eml"
			fout, err := os.Create(userFile)
			defer fout.Close()
			if err != nil {
				fmt.Println(userFile, err)
				return false, nil
			}
			fout.WriteString(data)
			var atts = mime.Attachments
			for i := 0; i < len(atts); i++ {
				att := atts[i]
				attFilePath := AttPath + numberStr + "," + att.FileName()
				attfout, err := os.Create(attFilePath)
				defer attfout.Close()
				if err != nil {
					fmt.Println(AttPath+att.FileName(), err)
					return false, nil
				}
				attfout.Write(att.Content())

				log.Printf("下载附件：", att.FileName())
			}

			return false, nil
		}); err != nil {
		log.Fatalf("%v\n", err)
	}
	log.Printf("备份完成,请查看e:\\emails文件夹")
	reader.ReadLine()
}

func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
