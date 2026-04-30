package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
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

func makePacket(Type int32, payload string) Packet {
	var rawPayload []byte = append([]byte(payload), '\x00', '\x00')
	return Packet{Length: 8 + int32(len(rawPayload)), Id: rand.Int31(), Type: Type, Payload: rawPayload}
}

func makeBuffer(packet *Packet) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, packet.Length)
	binary.Write(&buf, binary.LittleEndian, packet.Id)
	binary.Write(&buf, binary.LittleEndian, packet.Type)
	buf.Write(packet.Payload)
	return buf.Bytes()
}

func sendPacket(conn net.Conn, rawBytes []byte) error {
	_, err := conn.Write(rawBytes)
	if err != nil {
		return err
	}

	return nil
}

func readPacket(conn net.Conn) (Packet, error) {
	var length int32

	err := binary.Read(conn, binary.LittleEndian, &length)
	if err != nil {
		return Packet{}, err
	}

	var buf []byte = make([]byte, length)

	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return Packet{}, err
	}

	id := int32(binary.LittleEndian.Uint32(buf[0:4]))
	Type := int32(binary.LittleEndian.Uint32(buf[4:8]))
	payload := buf[8:]

	return Packet{Length: length, Id: id, Type: Type, Payload: payload}, err
}

func handleHelp() {
	var help string = "I've no idea either, good luck finding it."
	fmt.Println(help)
	os.Exit(0)
}

func handleAuthentication(conn net.Conn) error {
	var packet Packet = makePacket(LOGIN, os.Getenv("RCON_PASSWD"))

	var rawBytes []byte = makeBuffer(&packet)

	err := sendPacket(conn, rawBytes)
	if err != nil {
		return err
	}

	packet, err = readPacket(conn)
	if err != nil {
		return err
	}

	if packet.Id == -1 {
		return errors.New("wrong password")
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
		fmt.Println("Authentication falied:", error)
		os.Exit(1)
	}
}
