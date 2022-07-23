package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// This is the seed for cipher, len(seed) == len(key) == 16
// You can change this and get your own specified pwbook
const seed = "zj=bio3lax4q^mo5"

var (
	mainbook = book{
		nil, // left for CTRcrypt function to judge if the key has been set
		make([]pwItem, 0),
	}
	isChanged bool
)

// This struct is for gob encoding, the field should be capitalized
// singleton mode
type book struct {
	// If CRTcript(right) == seed, the password is right
	Right []byte
	// All contents
	BookItems []pwItem
}

// item of password book
type pwItem struct {
	Id          int
	Description string
	Password    []byte
}

func main() {
	// load file and prepare for reading command line
	fmt.Println("Pwbook, a password book written in Go.")
	if len(os.Args) != 2 {
		fmt.Println("Invalid command! Usage:\n    pwbook <filename>")
		return
	}
	load_file(os.Args[1])

	command_line_reader := bufio.NewReader(os.Stdin)
	var commands []string

	// command line loop
	for goon := false; !goon; goon = execute_line(commands) {
		fmt.Print(">>> ")
		text, err := command_line_reader.ReadString('\n')
		if err != nil {
			log.Fatal(fmt.Errorf("fatal error in reading command: %w", err))
		}
		// The last character "\n" should be removed
		text = text[:len(text)-1]
		commands = strings.Split(text, " ")
	}

	// save file when changed and confirmed
	if !isChanged {
		fmt.Println("Bye.")
		return
	}

	fmt.Println("Would you want to save changes?\n" +
		"Input \"y\" then enter to save, otherwise withdraw.")
	text, err := command_line_reader.ReadString('\n')
	if err != nil {
		log.Fatal(fmt.Errorf("fatal error in reading command: %w", err))
	}

	// comfirm whether save
	if text == "y\n" {
		save_file(os.Args[1])
		fmt.Println("Changes have been saved.\nBye.")
	} else {
		fmt.Println("Changes have been withdrawed.\nBye.")
	}
}

func load_file(filename string) {
	// check file state
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		fmt.Printf("File %s didn't exist. A new file will be "+
			"created when saving.\n", filename)
		mainbook.Right = CTRcrypt([]byte(seed))
		isChanged = true
		return
	}
	if err != nil {
		log.Fatal(fmt.Errorf("open file %s error: %w", filename, err))
	}

	// read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(fmt.Errorf("read file %s error: %w", filename, err))
	}
	buf := bytes.NewBuffer(data)

	// load the contents
	decoder := gob.NewDecoder(buf)
	err = decoder.Decode(&mainbook)
	if err != nil {
		log.Fatal(fmt.Errorf("decode file %s error: %w", filename, err))
	}
}

func save_file(filename string) {
	// encode contents
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(mainbook)
	if err != nil {
		log.Fatal(fmt.Errorf("encode file %s error: %w", filename, err))
	}

	// write file
	err = os.WriteFile(filename, buf.Bytes(), 0644)
	if err != nil {
		log.Fatal(fmt.Errorf("write file %s error: %w", filename, err))
	}
}

func execute_line(commands []string) bool {
	itempos, nok := is_bad_command(commands)
	if nok {
		return false
	}

	switch commands[0] {
	case "add":
		// This is to find a minimal id available
		idlist := make([]bool, len(mainbook.BookItems)+1)
		for _, cont := range mainbook.BookItems {
			if cont.Id <= len(mainbook.BookItems)+1 {
				idlist[cont.Id] = true
			}
		}
		var aid int
		for ; idlist[aid]; aid++ {
		}
		mainbook.BookItems = append(mainbook.BookItems, pwItem{aid, commands[1],
			CTRcrypt([]byte(commands[2]))})
		isChanged = true
		fmt.Printf("Item %s have been added.\n", commands[1])

	case "change":
		mainbook.BookItems[itempos].Password = CTRcrypt([]byte(commands[2]))
		isChanged = true
		fmt.Printf("Password of %s have been changed.\n",
			mainbook.BookItems[itempos].Description)

	case "get":
		fmt.Println("Password:   ", string(CTRcrypt(mainbook.BookItems[itempos].Password)))
		fmt.Println("Desription: ", mainbook.BookItems[itempos].Description)

	case "remove":
		// If no item exist, just remove it
		desp := mainbook.BookItems[itempos].Description
		mainbook.BookItems[itempos] = mainbook.BookItems[len(mainbook.BookItems)-1]
		mainbook.BookItems = mainbook.BookItems[:len(mainbook.BookItems)-1]
		isChanged = true
		fmt.Printf("Item %s have been removed.\n", desp)

	case "list":
		for _, i := range mainbook.BookItems {
			fmt.Println(i.Id, "\t", i.Description)
		}

	case "exit":
		return true
	}

	return false
}

// Check whether the command is right and get the pointer of pwItem chosed
func is_bad_command(commands []string) (itemsubs int, nok bool) {
	switch len(commands) {
	case 1:
		switch commands[0] {
		case "exit":
			return
		case "list":
			return
		// If nothing inputed, skip.
		case "":
			return 0, true
		case "help":
			goto USAGE
		}

	case 2:
		switch commands[0] {
		case "get":
			goto GETNUMBERID
		case "remove":
			goto GETNUMBERID
		}

	case 3:
		switch commands[0] {
		case "add":
			return
		case "change":
			goto GETNUMBERID
		}
	}

	// If command not match any, it is a bad command.
BADCOMMAND:
	fmt.Println("Unrecognized command.")

USAGE:
	fmt.Println(`Usage:
	exit                       exit and ask you to save or not
	help                       show this page
	list                       list all passwords' descriptions and id
	get <id>                   get password and description of <id>
	remove <id>                remove passwords of <id>
	add <describe> <password>  add password with description <describe>
	change <id> <password>     change password of <id>`)
	return 0, true

GETNUMBERID:
	// check if id is a integer
	id, err := strconv.Atoi(commands[1])
	if err != nil {
		goto BADCOMMAND
	}

	// find the id of pwItem to choose
	for i, cont := range mainbook.BookItems {
		if cont.Id == id {
			return i, false
		}
	}
	fmt.Printf("Identifier %d does not exist.\n", id)
	return 0, true
}

// encode and decode using CRT (same function)
func CTRcrypt(src []byte) (dst []byte) {
RETRY:
	// get the passwrod you input for authorize
	fmt.Println("[authority] Password:")

	// term package is to avoid password shown on the command line.
	key, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(fmt.Errorf("read password error:%w", err))
	}
	if len(key) != 16 {
		fmt.Println("You have inputed a wrong password!")
		goto RETRY
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(fmt.Errorf("create cipher error: %w", err))
	}

	stream := cipher.NewCTR(block, []byte(seed))

	// judge authority
	dst = make([]byte, 16)
	stream.XORKeyStream(dst, mainbook.Right)
	if mainbook.Right != nil && string(dst) != string(seed) {
		fmt.Println("You have inputed a wrong password!")
		goto RETRY
	}

	// crypt
	dst = make([]byte, len(src))
	stream.XORKeyStream(dst, src)
	return
}
