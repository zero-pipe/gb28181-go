package mansrtsp

import "fmt"

// Pause builds a MANSRTSP PAUSE body (PauseTime=0, matching common device captures).
func Pause(cseq int) string {
	return PauseAt(cseq, "0")
}

// PauseNow builds PAUSE with PauseTime: now.
func PauseNow(cseq int) string {
	return PauseAt(cseq, "now")
}

// PauseAt builds PAUSE with an explicit PauseTime value.
func PauseAt(cseq int, pauseTime string) string {
	if pauseTime == "" {
		pauseTime = "0"
	}
	return fmt.Sprintf("PAUSE RTSP/1.0\r\nCSeq: %d\r\nPauseTime: %s\r\n", cseq, pauseTime)
}

// Resume builds a MANSRTSP PLAY (resume from now) body.
func Resume(cseq int) string {
	return fmt.Sprintf("PLAY RTSP/1.0\r\nCSeq: %d\r\nRange: npt=now-\r\n", cseq)
}

// Speed builds a MANSRTSP PLAY with Scale.
func Speed(cseq int, speed float64) string {
	return fmt.Sprintf("PLAY RTSP/1.0\r\nCSeq: %d\r\nScale: %.6f\r\n", cseq, speed)
}

// Seek builds a MANSRTSP PLAY with npt seek offset (seconds).
func Seek(cseq int, seekTime int64) string {
	return fmt.Sprintf("PLAY RTSP/1.0\r\nCSeq: %d\r\nRange: npt=%d-\r\n", cseq, seekTime)
}

// SeekSpeed builds PLAY with both Scale and Range (抓包定位常合并倍速+seek).
func SeekSpeed(cseq int, seekTime int64, speed float64) string {
	return fmt.Sprintf("PLAY RTSP/1.0\r\nCSeq: %d\r\nScale: %.6f\r\nRange: npt=%d-\r\n", cseq, speed, seekTime)
}
