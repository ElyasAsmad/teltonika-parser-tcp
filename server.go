package main

import (
	"encoding/hex"
	"encoding/json"
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

// println("Altitude:", decoded.Data[i].Altitude)
// println("Angle:", decoded.Data[i].Angle)
// println("EventID:", decoded.Data[i].EventID)
// println("Lat:", decoded.Data[i].Lat)
// println("Lng:", decoded.Data[i].Lng)
// println("Priority:", decoded.Data[i].Priority)
// println("Speed:", decoded.Data[i].Speed)
// println("Utime UNIX:", decoded.Data[i].UtimeMs)
// println("Satellites:", decoded.Data[i].VisSat)

type DeviceData struct {
	IMEI     string      `json:"imei"`
	CodecID  string      `json:"codec_id"`
	NoOfData int         `json:"data_len"`
	Altitude int         `json:"altitude"`
	Angle    int         `json:"angle"`
	EventID  int         `json:"event_id"`
	Lat      int         `json:"lat"`
	Lng      int         `json:"lng"`
	Priority int         `json:"priority"`
	Speed    int         `json:"speed"`
	UtimeMs  int         `json:"utime_ms"`
	VisSat   int         `json:"vis_sat"`
	Elements []IOElement `json:"elements"`
}

type IOElement struct {
	Length int `json:"length"`
	IOID   int `json:"io_id"`
	Value  int `json:"value"`
}

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
					println("Angle:", decoded.Data[i].Angle)
					println("EventID:", decoded.Data[i].EventID)
					println("Lat:", decoded.Data[i].Lat)
					println("Lng:", decoded.Data[i].Lng)
					println("Priority:", decoded.Data[i].Priority)
					println("Speed:", decoded.Data[i].Speed)
					println("Utime UNIX:", decoded.Data[i].UtimeMs)
					println("Satellites:", decoded.Data[i].VisSat)

					elements := make([]IOElement, len(decoded.Data[i].Elements))

					for j := 0; j < len(decoded.Data[i].Elements); j++ {
						println("Element ID:", decoded.Data[i].Elements[j].IOID)
						println("Element len:", decoded.Data[i].Elements[j].Length)

						b2h := hex.EncodeToString(decoded.Data[i].Elements[j].Value)

						n := new(big.Int)

						n.SetString(b2h, 16)

						println("Element value:", n.Int64())
						print("\n")

						elements = append(elements, IOElement{
							Length: int(decoded.Data[i].Elements[j].Length),
							IOID:   int(decoded.Data[i].Elements[j].IOID),
							Value:  int(n.Int64()),
						})
					}

					el := DeviceData{
						IMEI:     imei,
						CodecID:  hex.EncodeToString([]byte{decoded.CodecID}),
						NoOfData: int(decoded.NoOfData),
						Altitude: int(decoded.Data[i].Altitude),
						Angle:    int(decoded.Data[i].Angle),
						EventID:  int(decoded.Data[i].EventID),
						Lat:      int(decoded.Data[i].Lat),
						Lng:      int(decoded.Data[i].Lng),
						Priority: int(decoded.Data[i].Priority),
						Speed:    int(decoded.Data[i].Speed),
						UtimeMs:  int(decoded.Data[i].UtimeMs),
						VisSat:   int(decoded.Data[i].VisSat),
						Elements: elements,
					}

					jsonData, err := json.Marshal(el)

					if err != nil {
						fmt.Println("Error while marshaling data", err)
						break
					}

					PublishToPubSub(jsonData)

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
