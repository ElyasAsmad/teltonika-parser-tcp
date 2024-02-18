package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net"

	"github.com/filipkroca/b2n"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "4554"
	CONN_TYPE = "tcp"
)

func main() {

	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}

	defer l.Close()

	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			return
		}

		go handleRequest(conn)
	}

}

func handleRequest(conn net.Conn) {
	var b []byte
	var imei string
	knownIMEI := true
	step := 1

	// Close the connection when you're done with it.
	defer conn.Close()

	for {
		// Make a buffer to hold incoming data.
		buf := make([]byte, 2048)

		// Read the incoming connection into the buffer.
		size, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}

		// Send a response if known IMEI and matches IMEI size
		// TODO: Add IMEI validation from Redis
		if knownIMEI {
			b = []byte{1} // 0x01 if we accept the message

			message := hex.EncodeToString(buf[:size])

			fmt.Println("----------------------------------------")
			fmt.Println("Data From:", conn.RemoteAddr().String())
			fmt.Println("Size of message: ", size)
			fmt.Println("Message:", message)
			fmt.Println("Step:", step)

			switch step {
			case 1:
				step = 2
				var imeiLen uint8 = (buf)[1]

				imei, err = b2n.ParseIMEI(&buf, 2, int(imeiLen))

				if err != nil {
					fmt.Printf("Decode error, %v", err)
					return
				}

				conn.Write(b)

			case 2:
				// elements, err := parseData(buf, size, imei)
				// if err != nil {
				// 	fmt.Println("Error while parsing data", err)
				// 	break
				// }

				// for i := 0; i < len(elements); i++ {
				// 	element := elements[i]
				// 	err := rc.Insert(&element)
				// 	if err != nil {
				// 		fmt.Println("Error inserting element to database", err)
				// 	}
				// }

				// conn.Write([]byte{0, 0, 0, uint8(len(elements))})

				decoded, err := DecodeAVL(&buf)

				if err != nil {
					fmt.Println("Error while parsing data", err)
					break
				}

				println("IMEI:", imei)
				println("Data len:", decoded.NoOfData)
				println("Codec:", hex.EncodeToString([]byte{decoded.CodecID}))

				for i := 0; i < int(decoded.NoOfData); i++ {

					println("Altitude:", decoded.Data[i].Altitude)
					println("Direction:", decoded.Data[i].Angle)
					println("Satellites:", decoded.Data[i].VisSat)
					println("Speed:", decoded.Data[i].Speed)
					println("EventID:", decoded.Data[i].EventID)
					println("Lat:", decoded.Data[i].Lat)
					println("Lng:", decoded.Data[i].Lng)

					for j := 0; j < len(decoded.Data[i].Elements); j++ {
						println("Element ID:", decoded.Data[i].Elements[j].IOID)
						println("Element len:", decoded.Data[i].Elements[j].Length)

						b2h := hex.EncodeToString(decoded.Data[i].Elements[j].Value)

						n := new(big.Int)

						n.SetString(b2h, 16)

						println("Element value:", n.Int64())
						print("\n")
					}

				}

				conn.Write([]byte{0, 0, 0, uint8(int(decoded.NoOfData))})
			}

		} else {
			b = []byte{0} // 0x00 if we decline the message

			conn.Write(b)
			break
		}
	}
}
