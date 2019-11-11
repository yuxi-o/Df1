package main

import (
	"fmt"
	"os"
	"strconv"

    "github.com/tarm/serial"
)
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
	fmt.Printf("Usage: %s /dev/ttyxxx mode speed databits parity stopbits\n", os.Args[0]);
	fmt.Printf("mode: specify the mode , full or half, not using, just for the same.\n");
	fmt.Printf("speed: specify the bps, 115200, 57600, 9600, 4800, 2400...\n");
	fmt.Printf("databits: 7 or 8")
	fmt.Printf("parity: 0:none, 1:even, 2:odd\n");
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
	case "1":
		sp = 'E'
	case "2":
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
			if rxbuf[index+8] == 0xA2 { // FNC: A2, to read
				count = int(rxbuf[index+9])
				for i =0; i < count; i++ { // the count of bytes to read
					txbuf[8+i] = byte(0x1 + i)
				}
				num = 6 + count
			} else if (rxbuf[index+8] == 0xAA) || (rxbuf[index+8] == 0xAB) { //FNC: AA, to write
				num = 6
			} else {
				fmt.Println("Not support FNC: 0x%.2X\n", rxbuf[index+8])
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
			fmt.Println("--------------------")
			break
		}
	}
}
