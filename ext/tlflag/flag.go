package tlflag

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nikandfor/errors"
	"github.com/nikandfor/tlog"
	"github.com/nikandfor/tlog/compress"
	"github.com/nikandfor/tlog/convert"
	"github.com/nikandfor/tlog/rotated"
)

var (
	OpenFileWriter = func(n string, f int, m os.FileMode) (io.Writer, error) {
		return os.OpenFile(n, f, m)
	}

	OpenFileReader = func(n string, f int, m os.FileMode) (io.Reader, error) {
		return os.OpenFile(n, f, m)
	}

	CompressorBlockSize = compress.MB

	JSONDefaultTimeFormat = "2006-01-02T15:04:05.999999Z07:00"
)

func updateFileFlags(of int, s string) int {
	for _, c := range s {
		if c == '0' {
			of |= os.O_TRUNC
		}
	}

	return of
}

func updateConsoleFlags(ff int, s string) int {
	for _, c := range s {
		switch c {
		case 'd':
			ff |= tlog.LdetFlags
		case 'm':
			ff |= tlog.Lmilliseconds
		case 'M':
			ff |= tlog.Lmicroseconds
		case 'n':
			ff |= tlog.Lfuncname
		case 'f':
			ff &^= tlog.Llongfile
			ff |= tlog.Lshortfile
		case 'F':
			ff &^= tlog.Lshortfile
			ff |= tlog.Llongfile
		case 'U':
			ff |= tlog.LUTC
		}
	}

	return ff
}

func updateConsoleOptions(w *tlog.ConsoleWriter, s string) {
	for i := 0; i < len(s); i++ {
	out:
		switch s[i] { //nolint:gocritic
		case 's':
			w.IDWidth = len(tlog.ID{})

			if i+1 == len(s) {
				break
			}

			cl := byte(')')
			switch s[i+1] {
			case '[', '{':
				cl = s[i+1] + 2
			case '(':
			default:
				break out
			}

			i += 2
			p := strings.IndexByte(s[i:], cl)
			if p == -1 {
				break
			}

			width, err := strconv.ParseInt(string(s[i:i+p]), 10, 8)
			if err == nil {
				w.IDWidth = int(width)
			}

			i += p
		case 'S':
			w.IDWidth = 2 * len(tlog.ID{})
		case 'c':
			w.Colorize = true
		case 'C':
			w.Colorize = false
		}
	}
}

func updateJSONOptions(w *convert.JSON, s string) {
	for i := 0; i < len(s); i++ {
	out:
		switch s[i] { //nolint:gocritic
		case 'L':
			w.AttachLabels = true
		case 'U':
			w.TimeInUTC = true
		case 't':
			w.TimeFormat = ""
		case 'T':
			w.TimeFormat = JSONDefaultTimeFormat

			if i+1 == len(s) {
				break
			}

			cl := byte(')')
			switch s[i+1] {
			case '[', '{':
				cl = s[i+1] + 2
			case '(':
			default:
				break out
			}

			i += 2
			p := strings.IndexByte(s[i:], cl)
			if p == -1 {
				break
			}

			w.TimeFormat = s[i : i+p]

			i += p
		}
	}
}

func OpenWriter(dst string) (wc io.WriteCloser, err error) {
	var ws tlog.TeeWriter

	defer func() {
		if err == nil {
			return
		}

		_ = ws.Close()
	}()

	for _, d := range strings.Split(dst, ",") {
		if d == "" {
			continue
		}

		var opts string
		//	of := os.O_APPEND | os.O_WRONLY | os.O_CREATE
		if p := strings.IndexByte(d, ':'); p != -1 {
			opts = d[p+1:]
			d = d[:p]

			//		ff, of = updateFlags(ff, of, opts)
		}

		var w io.Writer
		w, err = openw(d, opts)
		if err != nil {
			return nil, errors.Wrap(err, "%v", d)
		}

		ws = append(ws, w)
	}

	if len(ws) == 1 {
		var ok bool
		if wc, ok = ws[0].(io.WriteCloser); ok {
			return wc, nil
		}
	}

	return ws, nil
}

func openw(fn string, opts string) (wc io.Writer, err error) {
	// r = openFile(fn)
	// r = newDecompressor(r)
	// r = newJSONReader(r)
	// read r

	// w = openFile(fn)
	// w = newCompressor(w)
	// w = newJSONWriter(w)
	// write w

	of := os.O_APPEND | os.O_WRONLY | os.O_CREATE
	of = updateFileFlags(of, opts)

	mode := os.FileMode(0644)

	var w io.Writer
	var c io.Closer

	fmt := fn
	for { // pop extensions to find out if it's a file or stderr
		switch ext := filepath.Ext(fmt); ext {
		case ".tlog", ".tl", ".dump", ".log", "", ".json":
			switch strings.TrimSuffix(fmt, ext) {
			case "", "stderr":
				w = tlog.Stderr
			case "-", "stdout":
				w = tlog.Stdout
			default:
				if strings.ContainsRune(fn, rotated.SubstChar) {
					f := rotated.Create(fn)
					f.Flags = of
					f.MaxSize = 128 * rotated.MB

					w = f
					c = f
				} else {
					w, err = OpenFileWriter(fn, of, mode)

					c, _ = w.(io.Closer)
				}
			}
		case ".ez":
			fmt = strings.TrimSuffix(fmt, ext)

			continue
		default:
			err = errors.New("unsupported file ext: %v", ext)
		}

		break
	}
	if err != nil {
		return nil, err
	}

	fmt = fn
loop2:
	for {
		ext := filepath.Ext(fmt)

		switch ext {
		case ".tlog", ".tl":
		case ".dump":
			w = tlog.NewDumper(w)
		case ".ez":
			w = compress.NewEncoder(w, CompressorBlockSize)
		case ".log", "":
			ff := tlog.LstdFlags
			ff = updateConsoleFlags(ff, opts)

			cw := tlog.NewConsoleWriter(w, ff)

			updateConsoleOptions(cw, opts)

			w = cw
		case ".json":
			jw := convert.NewJSONWriter(w)

			updateJSONOptions(jw, opts)

			w = jw
		default:
			panic("missed extension switch case")
		}

		switch ext {
		case ".ez":
		default:
			break loop2
		}

		fmt = strings.TrimSuffix(fmt, ext)
	}

	if c != nil && w.(interface{}) != c.(interface{}) {
		return tlog.WriteCloser{
			Writer: w,
			Closer: c,
		}, nil
	}

	if c == nil {
		if _, ok := w.(io.Closer); ok {
			return tlog.NopCloser{
				Writer: w,
			}, nil
		}
	}

	return w, nil
}

func OpenReader(src string) (rc io.ReadCloser, err error) {
	r, err := openr(src)
	if err != nil {
		return nil, err
	}

	var ok bool
	if rc, ok = r.(io.ReadCloser); ok {
		return rc, nil
	}

	return tlog.NopCloser{
		Reader: r,
	}, nil
}

func openr(fn string) (rc io.Reader, err error) {
	var r io.Reader
	var c io.Closer

	fmt := fn
	for {
		switch ext := filepath.Ext(fmt); ext {
		case ".tlog", ".tl", "":
			switch strings.TrimSuffix(fmt, ext) {
			case "", "-", "stdin":
				r = tlog.Stdin
			default:
				r, err = OpenFileReader(fn, os.O_RDONLY, 0)
				c, _ = r.(io.Closer)
			}
		case ".ez":
			fmt = strings.TrimSuffix(fmt, ext)

			continue
		default:
			err = errors.New("unsupported file ext: %v", ext)
		}

		break
	}
	if err != nil {
		return
	}

	if ext := filepath.Ext(fn); ext == ".ez" {
		r = compress.NewDecoder(r)
	}

	if c != nil && r.(interface{}) != c.(interface{}) {
		return tlog.ReadCloser{
			Reader: r,
			Closer: c,
		}, nil
	}

	if c == nil {
		if _, ok := r.(io.Closer); ok {
			return tlog.NopCloser{
				Reader: r,
			}, nil
		}
	}

	return r, nil
}
