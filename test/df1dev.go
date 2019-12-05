package main

import (
	"fmt"
	"os"
	"strconv"
	"io"
	"bytes"
	"encoding/binary"

    "github.com/tarm/serial"
)

func encodeValue(value interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, value)
	return buf.Bytes(), err 
}

func decodeValue(reader io.Reader, value interface{}) error {
	err := binary.Read(reader, binary.LittleEndian, value)
	return err
}

func float32Value(buf []byte) (value float32, err error) {
	err = decodeValue(bytes.NewReader(buf), &value)
	return value, err
}

func int16Value(buf []byte) (value int16, err error) {
	err = decodeValue(bytes.NewReader(buf), &value)
	return value, err
}

func uint16Value(buf []byte) (value uint16, err error) {
	err = decodeValue(bytes.NewReader(buf), &value)
	return value, err
}

func calc_crc(crc uint16,  buf byte)( uint16){
	var temp, y uint16
	temp = crc ^ uint16(buf)
	crc = (crc & 0xff00) | (temp & 0xff)
	for y = 0; y < 8; y++ {
		if (crc & 1) == 1{
			crc = crc >> 1
			crc ^= 0xA001
		} else {
			crc = crc >> 1
		}
	} 
	return crc
}

func compute_crc(buf []byte, leng int)(crc uint16){
	var x int

	crc = 0
	for x = 0; x < leng; x++ {
		crc = calc_crc(crc, buf[x])
	}
	crc = calc_crc(crc, 0x03)
	return crc
}

func main(){
	fmt.Printf("[v1.1.2]Usage: %s /dev/ttyxxx mode speed databits parity stopbits\n", os.Args[0]);
	fmt.Printf("mode: specify the mode , full or half, not using, just for the same.\n");
	fmt.Printf("speed: specify the bps, 115200, 57600, 9600, 4800, 2400...\n");
	fmt.Printf("databits: 7 or 8")
	fmt.Printf("parity: 0:none, 1:odd, 2:even \n");
	fmt.Printf("stopbits: 1 or 2\n");
	fmt.Println("Parameters:", os.Args[1:])

	if 7 != len(os.Args) {
		fmt.Println("CMD error! Please repeat!\n")
		os.Exit(-1)
	}


	var sp byte
	speed, _ := strconv.Atoi(os.Args[3])
	sdb,_ := strconv.Atoi(os.Args[4])
	switch os.Args[5] {
	case "0":
		sp = 'N'
	case "2":
		sp = 'E'
	case "1":
		sp = 'O'
	default:
		sp = 'N'
	}
	ssb, _ := strconv.Atoi(os.Args[6])


	c := &serial.Config{Name: os.Args[1], Baud: speed, Size: byte(sdb), Parity: serial.Parity(sp), StopBits: serial.StopBits(ssb)}
	s, err := serial.OpenPort(c)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer s.Close()

	var N7arr [200] int16
	var F8arr [200] float32
	var B3arr [200] uint16 
	var binbuf []byte
	var s16 int16
	var f32 float32
	var i, u16 uint16

	for i = 0; i<200; i++ {
		N7arr[i] = 100 + int16(i)
		F8arr[i] = 100 + float32(i) + float32(i)*0.01
		B3arr[i] = 0x5555
	}
	
	txlen := 128
	rxbuf := make([]byte, 512)
	txbuf := make([]byte, txlen)
	for {
		sum := 0
		for {
			n, err := s.Read(rxbuf[sum:])
			if err != nil {
				fmt.Println(err)
				return

			} 

			sum += n
			if (sum > 4) && (rxbuf[sum-4] == 0x10) && (rxbuf[sum-3] == 0x03){
				break
			}
		}
		fmt.Printf("Receive %d bytes:\n%q\n", sum, rxbuf[:sum])
		
		var i, num, count int
		var crc uint16
		var id uint8
		minlen := sum - 10
		index := 0
		for index < minlen {
			for rxbuf[index] != 0x10 {
				index++	
			}
			if rxbuf[index+1] != 0x02 {
				index++
				continue
			}
			
			txbuf[txlen-2] = 0x10
			txbuf[txlen-1] = 0x06 
			s.Write(txbuf[txlen-2:txlen])
			
			txbuf[0] = 0x10
			txbuf[1] = 0x02
			txbuf[2] = rxbuf[index+3]
			txbuf[3] = rxbuf[index+2]
			if rxbuf[index+4] == 0x0F {
				txbuf[4] = 0x4F
			} else {
				fmt.Println("Not support CMD: 0x%.2X\n", rxbuf[index+4])
				break
			}
			txbuf[5] = rxbuf[index+5]
			txbuf[6] = rxbuf[index+6]
			txbuf[7] = rxbuf[index+7]
			count = int(rxbuf[index+9])
			id = rxbuf[index+12]
			if rxbuf[index+8] == 0xA2 { // FNC: A2, to read
				if id < 200 && count <= 4 {
					if rxbuf[index+10] == 0x07 && rxbuf[index+11] == 0x89 {
						binbuf, err = encodeValue(N7arr[id])	
						if err == nil {
							txbuf[8] = binbuf[0]
							txbuf[9] = binbuf[1]
						} else {
							fmt.Printf("EncodeValue error: %d:%v", id, N7arr[id])
							txbuf[8] = 0x0; 
							txbuf[9] = 0x0;
						}
					} else if rxbuf[index+10] == 0x08 && rxbuf[index+11] == 0x8A {
						binbuf, err = encodeValue(F8arr[id])	
						if err == nil {
							txbuf[8] = binbuf[0]
							txbuf[9] = binbuf[1]
							txbuf[10] = binbuf[2]
							txbuf[11] = binbuf[3]
						} else {
							fmt.Printf("EncodeValue error: %d:%v", id, N7arr[id])
							txbuf[8] = 0x0; 
							txbuf[9] = 0x0;
							txbuf[10] = 0x0;
							txbuf[11] = 0x0;
						}
					} else if rxbuf[index+10] == 0x03 && rxbuf[index+11] == 0x85 {
						binbuf, err = encodeValue(B3arr[id])	
						if err == nil {
							txbuf[8] = binbuf[0]
							txbuf[9] = binbuf[1]
						} else {
							fmt.Printf("EncodeValue error: %d:%v", id, N7arr[id])
							txbuf[8] = 0x0; 
							txbuf[9] = 0x0;
						}
					}
				} else {
					txbuf[8] = 0; 
					txbuf[9] = 0;
					txbuf[10] = 0;
					txbuf[11] = 0;
					for i =4; i < count; i++ { // the count of bytes to read
						txbuf[8+i] = 0 
					}
				}
				num = 6 + count
			} else if (rxbuf[index+8] == 0xAA) { //FNC: AA, to write
				if id < 200 {
					if rxbuf[index+10] == 0x07 && rxbuf[index+11] == 0x89 {
						s16, err = int16Value(rxbuf[(index+14):(index+16)])	
						if err == nil {
							N7arr[id] = s16 
							fmt.Println(s16)
						} else {
							fmt.Printf("AA DecodeValue sint error: %v", rxbuf[(index+14):(index+16)])
						}
					} else if rxbuf[index+10] == 0x08 && rxbuf[index+11] == 0x8A {
						f32, err = float32Value(rxbuf[(index+14):(index+18)])	
						if err == nil {
							F8arr[id] = f32 
							fmt.Println(f32)
						} else {
							fmt.Printf("AA DecodeValue float32 error: %v", rxbuf[(index+14):(index+18)])
						}
					} else if rxbuf[index+10] == 0x03 && rxbuf[index+11] == 0x85 {
						u16, err = uint16Value(rxbuf[(index+14):(index+16)])	
						if err == nil {
							B3arr[id] = u16 
							fmt.Println(u16)
						} else {
							fmt.Printf("AA DecodeValue bool error: %v", rxbuf[(index+14):(index+16)])
						}
					}
				}

				num = 6
			} else if (rxbuf[index+8] == 0xAB) { //FNC: AB, to write
				if id < 200 {
					if rxbuf[index+10] == 0x03 && rxbuf[index+11] == 0x85 {
						u16, err = uint16Value(rxbuf[(index+14):(index+16)])	
						uu16, errr := uint16Value(rxbuf[(index+16):(index+18)])	
						if err == nil && errr == nil {
							B3arr[id] = (B3arr[id] & (^u16)) | uu16 
							fmt.Println(B3arr[id])
						} else {
							fmt.Printf("AB DecodeValue bool error: %v", rxbuf[(index+14):(index+18)])
						}
					}
				}
				num = 6
			} else {
				fmt.Printf("Not support FNC: 0x%.2X\n", rxbuf[index+8])
				break	
			}

			crc = compute_crc(txbuf[2:], num)
			txbuf[num+2] = 0x10
			txbuf[num+3] = 0x03
			txbuf[num+4] = byte(crc & 0xFF) 
			txbuf[num+5] = byte((crc>>8)&0xFF) 
			num = num + 6

			_, err = s.Write(txbuf[:num])
			if err != nil {
				fmt.Println(err)
				_, err = s.Write(txbuf[:num])
			}
			fmt.Printf("Send %d data:\n0x10 0x06\n%q\n", num+2, txbuf[:num])
			fmt.Printf("---------[5:0x%X] [6:0x%X] [7:0x%X]-----------\n", txbuf[5], txbuf[6], txbuf[7])
			break
		}
	}
}
