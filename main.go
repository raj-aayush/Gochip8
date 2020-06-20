package main

import (
	"fmt"
	"math/rand"
	"time"
	"bufio"
	"os"
	"pkg/filostack"
)

var blankScreen [32]uint64

type chip8 struct {
	mem [4000]uint8
	screen [32]uint64
	reg [16]uint8
	ind uint16
	pc uint16
	sp uint16
	soundTimer uint8
	delayTimer uint8
	st filoStack
	inputReader * bufio.Reader
}

func main() {

	var system chip8
	system.st.stack = []uint16{}
	system.inputReader = bufio.NewReader(os.Stdin)
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Booting up...")
	system.pc = 1


}

func (system *chip8) GetNextOp() uint16 {
	return (uint16(system.mem[system.pc]) << 8) + uint16(system.mem[system.pc + 1])
}

func (system *chip8) DrawSprite(x uint8, y uint8, h uint8) {
	for i := uint8(0); i < h; i++ {
		row := uint8(y+i % 32)
		if x > 64 - 8 {
			// Wrap sprite
			pixelsInBounds := 64-x
			pixelsOverflow := 8 - pixelsInBounds
			overFlowImg := uint8(system.mem[system.ind + uint16(i)] << pixelsInBounds)
			inBoundsImg := uint8(system.mem[system.ind + uint16(i)] >> pixelsOverflow)
			system.PasteSpriteLine(0, row, overFlowImg)
			system.PasteSpriteLine(64-8, row, inBoundsImg)
		} else {
			// Paste sprite
			system.PasteSpriteLine(x, row, system.mem[system.ind + uint16(i)])
		}
	}
}

func (system *chip8) PasteSpriteLine(vx uint8, vy uint8, img uint8) {
	rowMask := uint64((0xFF << (64 - vx)))
	rowVal  := uint8((system.screen[vy] & rowMask) >> (64 - vx))
	if rowVal ^ img != 0 {
		system.reg[0xF] = 1
	}
	rowVal ^= img
	rowVal = ^rowVal
	row := ^uint64(rowVal << (64-vx))
	system.screen[vy] &= row
}

func (system *chip8) SpriteAddr(char uint8) uint16 {
	return uint16((char & 0x000F) * 5)
}

func (system *chip8) ReadKey(input chan rune) {
    char, _, err := system.inputReader.ReadRune()
    if err != nil {
        fmt.Println(err)
    }
    input <- char
}

func extractReg(x int, opcode uint16) uint8 {
	mask := uint16(0xF) << x*4
	return uint8(opcode & mask >> x*4)
}

func (system *chip8) ExecOp(op uint16) {
	switch ins := (op & 0xF000) >> 12; ins {
	case 0x0:
		switch op {
		case 0x00E0:
			// Clear screen
			fmt.Println("00E0")
			system.screen = blankScreen
		case 0x00EE:
			// Return from subroutine
			fmt.Println("00EE")
			var err error
			system.pc, err = system.st.pop()
			if err != nil {
				fmt.Println(err)
			}
			system.sp--
		}
		system.pc += 2
	case 0x1:
		// 0x1NNN
		// jmp nnn
		// Jump to address at op-0x1000
		fmt.Println("1NNN")
		system.pc = uint16(op & 0x0FFF) & 0x0FFF
	case 0x2:
		// 0x2NNN
		// jsr nnn
		// Call subroutine at op-0x2000
		fmt.Println("2NNN")
		addr := uint16(op & 0x0FFF) & 0x0FFF
		system.st.push(system.pc)
		system.pc = addr
		system.sp++
	case 0x3:
		// 0x3XRR
		// skeq vx, rr
		// Skip next op if reg[vx] == rr
		fmt.Println("3XRR")
		vx := extractReg(2, op)
		rr := uint8(op & 0x00FF)
		if system.reg[vx] == rr {
			system.pc += 2
		}
		system.pc += 2
	case 0x4:
		// 0x4XRR
		// skne vx, rr
		// Skip next op if reg[vx] != rr
		fmt.Println("4XRR")
		vx := extractReg(2, op)
		rr := uint8(op & 0x00FF)
		if system.reg[vx] != rr {
			system.pc += 2
		}
		system.pc += 2
	case 0x5:
		// 0x5XY0
		// skeq vx, vy
		// Skip next op if reg[vx] == reg[vy]
		fmt.Println("5XY0")
		vx := extractReg(2, op)
		vy := extractReg(1, op)
		if system.reg[vx] == system.reg[vy] {
			system.pc += 2
		}
		system.pc += 2
	case 0x6:
		// 0x6XRR
		// mov vx, rr
		// reg[vx] = rr
		fmt.Println("6XRR")
		vx := extractReg(2, op)
		rr := uint8(op & 0x00FF)
		system.reg[vx] = rr
		system.pc += 2
	case 0x7:
		// 0x7XRR
		// add vx, rr
		// reg[vx] += rr
		// No carry is generated
		fmt.Println("7XRR")
		vx := extractReg(2, op)
		rr := uint8(op & 0x00FF)
		system.reg[vx] += rr
		system.pc += 2
	case 0x8:
		// 0x8000
		// Reg ops:
		vx := extractReg(2, op)
		vy := extractReg(1, op)
		switch subIns := op & 0x000F; subIns {
		case 0x0:
			// 0x8XY0
			// mov vx, vy
			// reg[vx] = reg[vy]
			fmt.Println("8XY0")
			system.reg[vx] = system.reg[vy]
		case 0x1:
			// 0x8XY1
			// or vx, vy
			// reg[vx] |= reg[vy]
			fmt.Println("8XY1")
			system.reg[vx] |= system.reg[vy]
		case 0x2:
			// 0x8XY2
			// and vx, vy
			// reg[vx] &= reg[vy]
			fmt.Println("8XY2")
			system.reg[vx] &= system.reg[vy]
		case 0x3:
			// 0x8XY3
			// xor vx, vy
			// reg[vx] ^= reg[vy]
			fmt.Println("8XY3")
			system.reg[vx] ^= system.reg[vy]
		case 0x4:
			// 0x8XY4
			// add vx, vy
			// reg[vx] += reg[vy]
			// reg[0xF] = carry
			fmt.Println("8XY4")
			sum32Bit := uint32(system.reg[vx]) + uint32(system.reg[vy])
			if sum32Bit > uint32(0xFFFF) {
				system.reg[0xF] = 1
			} else {
				system.reg[0xF] = 0
			}
			system.reg[vx] += system.reg[vy]
		case 0x5:
			// 0x8XY5
			// sub vx, vy
			// reg[vx] -= reg[vy]
			// if borrowed: reg[0xF] = 1
			fmt.Println("8XY5")
			if system.reg[vx] < system.reg[vy] {
				system.reg[0xF] = 0
			} else {
				system.reg[0xF] = 1
			}
			system.reg[vx] = system.reg[vx] - system.reg[vy]
		case 0x6:
			// 0x8X06
			// shr vx
			// reg[vr] = reg[vr] >> 1
			// reg[0xF] = 0
			fmt.Println("8X06")
			system.reg[0xF] = system.reg[vx] & 0x01
			system.reg[vx] >>=  1
		case 0x7:
			// 0x8XY7
			// rsb vx, vy
			// reg[vx] = reg[vy] - reg[vx]
			// if borrowed: reg[0xF] = 1
			fmt.Println("8XY7")
			if system.reg[vy] < system.reg[vx] {
				system.reg[0xF] = 0
			} else {
				system.reg[0xF] = 1
			}
			system.reg[vx] = system.reg[vy] - system.reg[vx]
		case 0xE:
			// 0x8X0E
			// shl vx
			// reg[vx] = reg[vx] << 1
			// reg[0xF] = 7
			fmt.Println("8X0E")
			system.reg[0xF] = (system.reg[vx] & 0x80) >> 7
			system.reg[vx] <<= 1
		}
		system.pc += 2
	case 0x9:
		// 0x9XY0
		// skne vx, vy
		// Skip next op if reg[vx] != reg[vy]
		fmt.Println("8XY0")
		vx := extractReg(2, op)
		vy := extractReg(1, op)
		if system.reg[vx] != system.reg[vy] {
			system.pc += 2
		}
		system.pc += 2
	case 0xA:
		// 0xANNN
		// mvi nnn
		// ind = nnn
		fmt.Println("ANNN")
		system.ind = op & 0x0FFF
		system.pc += 2
	case 0xB:
		// 0xBNNN
		// jmi nnn
		// Jump to address at nnn + reg[0x0]
		fmt.Println("BNNN")
		system.pc = (op & 0x0FFF) + uint16(system.reg[0])
	case 0xC:
		// 0xCXKK
		// rand vx, rr
		// reg[vx] = random() & rr
		fmt.Println("CNKK")
		vx := extractReg(2, op)
		rr := uint8(op & 0x00FF)
		system.reg[vx] = rr & uint8(rand.Uint32() & 0x000000FF)
		system.pc += 2
	case 0xD:
		// 0xDXYN
		// sprite vx, vy, n
		// Draw sprite at screen location (reg[vx], reg[vy]) height N
		// Sprite is stored in mem[ind], max 8bits wide
		// Wraps around the screen
		// If when drawn, clears a pixel, register VF is set to 1 otherwise 0.
		// All drawing is XOR drawing (toggle the screen pixels)
		fmt.Println("DXYN")
		vx := extractReg(2, op)
		vy := extractReg(1, op)
		height := uint8(op & 0x000F)
		system.DrawSprite(system.reg[vx], system.reg[vy], height)
		system.pc += 2
	case 0xE:
		switch subIns := op & 0x00FF; subIns {
		case 0x009E:
			// 0xEK9E
			// skpr k
			// skip if key (register rk) pressed
			fmt.Println("EK9E")
			issue()
		case 0x00A1:
			// 0xEKA1
			// skup k
			// skip if key (register rk) not pressed
			fmt.Println("EKA1")
			issue()
		}
	case 0xF:
		// 0xF000
		vr := extractReg(2, op)
		switch subIns := op & 0x00FF; subIns {
		case 0x07:
			// 0xFR07
			// gdelay vr
			// get delay timer into reg[vr]
			fmt.Println("FR07")
			system.reg[vr] = system.delayTimer
		case 0x0A:
			// 0xFR0A
			// key vr
			// Wait for keypress, put key in register vr
			fmt.Println("FR0A")
			input := make(chan rune, 1)
			go system.ReadKey(input)
			select {
				case i := <-input:
					system.reg[vr] = uint8(i - '0')
			}
		case 0x15:
			// 0xFR15
			// sdelay vr
			// set the delay timer to vr
			fmt.Println("FR15")
			system.delayTimer = system.reg[vr]
		case 0x18:
			// 0xFR18
			// ssound vr
			// set the sound timer to vr
			fmt.Println("FR18")
			system.soundTimer = system.reg[vr]
		case 0x1E:
			// 0xFR1E
			// adi vr
			// add register vr to the index register
			fmt.Println("FR1E")
			system.ind = (system.ind + uint16(system.reg[vr])) & 0x0FFF
		case 0x29:
			// 0xFR29
			// font vr
			// point I to the sprite for hexadecimal character in vr
			// Sprite is 5 bytes high
			fmt.Println("FR29")
			system.ind = system.SpriteAddr(system.reg[vr])
		case 0x33:
			// 0xFR33
			// Store BCD representation of register vr at location I, I+1, I+2
			// Don't change I
			fmt.Println("FR33")
			dec := system.reg[vr]
			for i := uint16(2); i >= 0; i-- {
				system.mem[system.ind+i] = uint8(dec % 10)
				dec = dec/10
			}
		case 0x55:
			// 0xFR55
			// str v0-vr
			// Store registers V0 - VR at location I onwards
			// I is incremeneted to opint to next location I = I + r + 1
			fmt.Println("FR55")
			for vi := uint16(0); vi <= uint16(vr); vi++ {
				system.mem[system.ind + vi] = system.reg[vi]
			}
		case 0x65:
			// 0xFR65
			// ldr v0-vr
			// load registers V0 - VR from locations I onwards
			// I is incremeneted to opint to next location I = I + r + 1
			fmt.Println("FR65")
			for vi := uint16(0); vi <= uint16(vr); vi++ {
				system.reg[vi] = system.mem[system.ind + vi]
			}
		}
		system.pc += 2
	}
}