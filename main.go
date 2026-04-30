package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
)

type Packet struct {
	Length  int32
	Id      int32
	Type    int32
	Payload []byte
}

const (
	MULTI_PACKET = 0
	COMMAND      = 2
	LOGIN        = 3
)

func handleHelp() {
	var help string = "I've no idea either, good luck finding it."
	fmt.Println(help)
	os.Exit(0)
}

func makeBuffer(packet *Packet) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, packet.Length)
	binary.Write(&buf, binary.LittleEndian, packet.Id)
	binary.Write(&buf, binary.LittleEndian, packet.Type)
	buf.Write(packet.Payload)
	return buf.Bytes()
}

func handleAuthentication(conn net.Conn) error {
	var pass []byte = append([]byte(os.Getenv("RCON_PASSWD")), '\x00', '\x00')
	var length int32 = 8 + int32(len(pass))

	var packet Packet = Packet{Length: length, Id: 7, Type: LOGIN, Payload: pass}

	var rawBytes []byte = makeBuffer(&packet)
	_, err := conn.Write(rawBytes)

	if err != nil {
		return err
	}

	var response [4096]byte
	n, err := conn.Read(response[:])

	if err != nil {
		return err
	}

	return nil
}

func handleCommand(conn net.Conn) error {
	var cmd []byte = append([]byte(strings.Join(os.Args[1:len(os.Args)], " ")), '\x00', '\x00')
	var length int32 = 8 + int32(len(cmd))

	var packet Packet = Packet{Length: length, Id: 3, Type: COMMAND, Payload: cmd}

	var rawBytes []byte = makeBuffer(&packet)

	_, err := conn.Write(rawBytes)

	if err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("What do you want?")
		fmt.Println("Try 'xxx --help' for more information.")
		os.Exit(1)
	}

	if os.Args[1] == "--help" {
		handleHelp()
	}

	var address string = os.Getenv("RCON_IP")

	conn, err := net.Dial("tcp", address)

	if err != nil {
		fmt.Println("Connect failed:", err)
		os.Exit(1)
	}

	defer conn.Close()

	if error := handleAuthentication(conn); error != nil {
		fmt.Println("Auth falied:", error)
		os.Exit(1)
	}

	if error := handleCommand(conn); error != nil {
		fmt.Println("What?")
		os.Exit(1)
	}

}
