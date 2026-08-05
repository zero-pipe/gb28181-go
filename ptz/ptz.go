package ptz

import (
	"fmt"
	"strings"
)

// Command builds a classic 8-byte PTZCmd hex string for direction control.
func Command(direction string, horizonSpeed, verticalSpeed, zoomSpeed int) string {
	if horizonSpeed <= 0 {
		horizonSpeed = 100
	}
	if verticalSpeed <= 0 {
		verticalSpeed = 100
	}
	if zoomSpeed <= 0 {
		zoomSpeed = 16
	}
	var cmd byte = 0x00
	switch strings.ToLower(direction) {
	case "left":
		cmd = 0x01
	case "right":
		cmd = 0x02
	case "up":
		cmd = 0x03
	case "down":
		cmd = 0x04
	case "upleft":
		cmd = 0x05
	case "upright":
		cmd = 0x06
	case "downleft":
		cmd = 0x07
	case "downright":
		cmd = 0x08
	case "zoomin":
		cmd = 0x09
	case "zoomout":
		cmd = 0x0a
	case "stop":
		cmd = 0x00
	default:
		cmd = 0x00
	}
	check := (0xA5 + 0x0F + 0x01 + int(cmd) + horizonSpeed + verticalSpeed + zoomSpeed) % 256
	return fmt.Sprintf("%02X%02X%02X%02X%02X%02X%02X%02X",
		0xA5, 0x0F, 0x01, cmd, byte(horizonSpeed), byte(verticalSpeed), byte(zoomSpeed), byte(check))
}

// FrontEndCmd builds an 8-byte front-end command hex string (preset set/call/delete, etc.).
func FrontEndCmd(cmdCode, parameter1, parameter2, combineCode2 int) string {
	b7 := (combineCode2 << 4) & 0xFF
	check := (0xA5 + 0x0F + 0x01 + cmdCode + parameter1 + parameter2 + b7) % 0x100
	return fmt.Sprintf("A50F01%02X%02X%02X%02X%02X",
		byte(cmdCode), byte(parameter1), byte(parameter2), byte(b7), byte(check))
}
