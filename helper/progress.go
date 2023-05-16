package helper

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/dustin/go-humanize"
	"github.com/gobackup/gobackup/logger"
	"github.com/hako/durafmt"
)

const (
	progressbarTemplate = `{{string . "time"}} {{string . "prefix"}}{{bar . "[" "=" "=" "-" "]"}} {{percent .}} ({{speed .}})`
)

type ProgressBar struct {
	bar        *pb.ProgressBar
	FileLength int64
	Reader     io.Reader
	logger     logger.Logger
	startTime  time.Time
}

func NewProgressBar(myLogger logger.Logger, reader *os.File) ProgressBar {
	info, _ := reader.Stat()
	fileLength := info.Size()

	bar := pb.ProgressBarTemplate(progressbarTemplate).Start64(fileLength)
	bar.SetWidth(100)
	bar.Set("time", time.Now().Format(logger.TimeFormat))
	bar.Set("prefix", myLogger.Prefix())

	multiReader := bar.NewProxyReader(reader)

	progressBar := ProgressBar{bar, fileLength, multiReader, myLogger, time.Now()}
	progressBar.start()

	return progressBar
}

func (p ProgressBar) start() {
	logger := p.logger

	logger.Infof("-> Uploading (%s)...", humanize.Bytes(uint64(p.FileLength)))
}

func (p ProgressBar) Errorf(format string, err ...any) error {
	p.bar.Finish()

	return fmt.Errorf(format, err...)
}

func (p ProgressBar) Done(url string) {
	logger := p.logger

	p.bar.Finish()
	t := time.Now()
	elapsed := t.Sub(p.startTime)

	logger.Info(fmt.Sprintf("Uploaded: %s (Duration %v)", url, durafmt.Parse(elapsed).LimitFirstN(2).String()))
}
