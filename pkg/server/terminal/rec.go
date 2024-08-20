package terminal

import (
	"bufio"
	"flutelake/fluteNAS/pkg/module/flog"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

// 同时写历史文件时的锁，防止并发写
var historyLock *sync.Mutex

type Recorder struct {
	uniqueID     string
	serialNumber string
	outFileName  string
	// 终端缓存
	outputFile *os.File

	// 历史命令
	historyFile *os.File
	nullFile    *os.File
	inr         *os.File
	inw         *os.File
	vt          *term.Terminal
}

const ReocrdFilePathPrefix = "/data/.kunlun-ssh-records"

func NewRecorder(serialNumber string, filename string) *Recorder {
	w := &Recorder{
		uniqueID:     filename,
		serialNumber: serialNumber,
		outFileName:  filepath.Join(ReocrdFilePathPrefix, filename),
	}
	if historyLock == nil {
		historyLock = &sync.Mutex{}
	}
	return w
}

func (w *Recorder) createRecFile() error {
	var err error
	w.outputFile, err = os.OpenFile(w.outFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0)
	if err != nil {
		return fmt.Errorf("create record file error: %v", err)
	}
	// 历史命令执行记录
	historyPath := filepath.Join(ReocrdFilePathPrefix, w.serialNumber)
	w.historyFile, err = os.OpenFile(historyPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0)
	if err != nil {
		return fmt.Errorf("create or open history file error: %v", err)
	}
	w.nullFile, err = os.Open(os.DevNull)
	if err != nil {
		return fmt.Errorf("open dev null error: %v", err)
	}
	w.inr, w.inw, err = os.Pipe()
	if err != nil {
		return fmt.Errorf("create pipe failed: %v", err)
	}
	screen := struct {
		io.Reader
		io.WriteCloser
	}{w.inr, w.nullFile}
	w.vt = term.NewTerminal(screen, "")

	return nil
}

func (w *Recorder) WriteData(data []byte) {
	if w.outputFile == nil {
		err := w.createRecFile()
		if err != nil {
			return
		}
	}
	timestamp := time.Now().Unix()
	fmt.Fprintf(
		w.outputFile,
		"%d,o,%s\n",
		timestamp,
		escapeNonPrintableChars(data),
	)
}

func (w *Recorder) WriteSize(rows int, cols int) error {
	if w.outputFile == nil {
		err := w.createRecFile()
		if err != nil {
			return err
		}
	}
	fmt.Fprintf(
		w.outputFile,
		"%d,r,%dx%d\n",
		time.Now().Unix(),
		cols,
		rows,
	)
	if w.vt != nil {
		w.vt.SetSize(rows, cols)
	}
	return nil
}

func (w *Recorder) Write(data []byte) (n int, err error) {
	w.WriteData(data)
	w.inw.Write(data)
	return len(data), err
}

func (w *Recorder) ParseCommandLine(str string) {
	for idx, r := range str {
		if r == '➜' {
			command := str[idx:]
			arr := strings.SplitN(command, " ", 3)
			if len(arr) != 3 || len(arr[2]) == 0 {
				break
			}
			w.saveCommand(command)
			// fflog.Printf("command: %s\n", command)
			break
		}
	}
}

func (w *Recorder) saveCommand(str string) (n int, err error) {
	if str == "" {
		return
	}
	historyLock.Lock()
	defer historyLock.Unlock()
	return w.historyFile.Write([]byte(fmt.Sprintf("%d,%s,%s\n", time.Now().Unix(), w.uniqueID, str)))
}

func (w *Recorder) ReadLine(handler func(bs []byte)) error {
	if w.outFileName == "" {
		return fmt.Errorf("rec file not exist")
	}
	f, err := os.Open(w.outFileName)
	if err != nil {
		return err
	}
	if f != nil {
		s := bufio.NewScanner(f)
		for s.Scan() {
			if s.Text() == "" {
				continue
			}

			arr := strings.SplitN(s.Text(), ",", 3)
			if len(arr) < 3 {
				continue
			}
			if arr[1] == "r" {
				continue
			}
			bs, _ := unescapePrintableChars(arr[2])
			handler(bs)
		}
	}
	return nil
}

func (w *Recorder) Close() {
	if w.historyFile != nil {
		w.historyFile.Close()
	}
	if w.outputFile != nil {
		w.outputFile.Close()
	}
	if w.inr != nil {
		w.inr.Close()
	}
	if w.inw != nil {
		w.inw.Close()
	}
	if w.nullFile != nil {
		w.nullFile.Close()
	}
}

// 转义不显示的字符
func escapeNonPrintableChars(data []byte) string {
	result := ""
	for i := 0; i < len(data); {
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size == 1 {
			// Handle non-printable characters
			result += fmt.Sprintf("\\u%04X", data[i])
			i++
		} else {
			// Escape special JSON characters
			if r == '"' {
				result += "\\\""
			} else if r == '\\' {
				result += "\\\\"
			} else if r < ' ' {
				// Handle other non-printable characters
				result += fmt.Sprintf("\\u%04X", r)
			} else {
				result += string(r)
			}
			i += size
		}
	}
	return result
}

// 恢复转义
func unescapePrintableChars(escapedStr string) ([]byte, error) {
	var result []byte
	for i := 0; i < len(escapedStr); {
		if escapedStr[i] == '\\' {
			if i+1 < len(escapedStr) && escapedStr[i+1] == 'u' {
				// Handle Unicode escape sequences
				if i+6 <= len(escapedStr) {
					hex := escapedStr[i+2 : i+6]
					r, err := strconv.ParseUint(hex, 16, 32)
					if err != nil {
						return nil, err
					}
					result = append(result, byte(r))
					i += 6
				} else {
					return nil, fmt.Errorf("invalid unicode escape sequence")
				}
			} else if i+1 < len(escapedStr) {
				// Handle escaped special characters
				nextChar := escapedStr[i+1]
				switch nextChar {
				case '"':
					result = append(result, '"')
				case '\\':
					result = append(result, '\\')
				default:
					return nil, fmt.Errorf("unsupported escape sequence")
				}
				i += 2
			} else {
				return nil, fmt.Errorf("incomplete escape sequence")
			}
		} else {
			result = append(result, escapedStr[i])
			i++
		}
	}
	return result, nil
}

// 过期天数 30天
const RecordExpireDay = 30

// 每天4点钟，清理一次历史缓存文件
func recCleanerGo() {
	now := time.Now()

	target := now.Truncate(4 * time.Hour)

	if now.After(target) {
		target = target.Add(time.Hour * 24)
	}

	// 创建定时器
	ticker := time.NewTicker(time.Until(target))
	<-ticker.C

	// 执行清理方法
	cleaner()

	// 重置定时器，等待下一个4点钟
	ticker.Reset(24 * time.Hour)
}

// 清理过期的缓存
func cleaner() {
	entries, err := os.ReadDir(ReocrdFilePathPrefix)
	if err != nil {
		flog.Errorf("ReadDir %s error: %v", ReocrdFilePathPrefix, err)
		return
	}

	if len(entries) == 0 {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := filepath.Join(ReocrdFilePathPrefix, entry.Name())
		cleanRecordFileContent(filePath)
	}
}

func cleanRecordFileContent(filePath string) {
	f, err := os.Open(filePath)
	if err != nil {
		flog.Errorf("ReadFile %s error: %v", filePath, err)
		return
	}
	defer f.Close()

	// 创建一个临时文件
	tmpFilePath := filepath.Join(ReocrdFilePathPrefix, "tmp_record")
	tmp, err := os.Create(tmpFilePath)
	if err != nil {
		flog.Errorf("create record file error: %v", err)
		return
	}
	defer tmp.Close()
	length := 0
	s := bufio.NewScanner(f)

	keepTs := time.Now().Add(-time.Hour * 24 * RecordExpireDay).Unix()
	for s.Scan() {
		if s.Text() == "" {
			continue
		}

		arr := strings.SplitN(s.Text(), ",", 3)
		if len(arr) < 3 {
			continue
		}
		ts := arr[0]
		t, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			flog.Errorf("convert string to int64 timestamp error: %v", err)
			continue
		}
		if t < keepTs {
			continue
		}

		n, _ := tmp.Write(s.Bytes())
		length = length + n
	}
	tmp.Close()
	if length > 0 {
		// 临时文件覆盖原文件
		err = os.Rename(tmpFilePath, filePath)
		if err != nil {
			flog.Errorf("move cleaned file error: %v", err)
		}
	} else {
		// 原文件的内容已经完全过期，删除原文件
		err = os.Remove(filePath)
		if err != nil {
			flog.Errorf("move expired file error: %v", err)
		}
	}
}
