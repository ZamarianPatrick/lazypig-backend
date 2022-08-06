package sensors

import (
	"encoding/binary"
	"fmt"
	"log"
	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
	"time"
)

func main() {

	state, err := host.Init()

	bus, _ := i2creg.Open("1")

	dev := i2c.Dev{
		Bus:  bus,
		Addr: 0x08,
	}

	var _ conn.Conn = &dev

	for true {
		write := []byte{0x20 + 0}
		read := make([]byte, 2)
		if err := dev.Tx(write, read); err != nil {
			log.Fatalln("here:", err)
		}

		fmt.Println(binary.LittleEndian.Uint16(read))

		time.Sleep(time.Millisecond * 200)
	}

	fmt.Println(state, err)
}
