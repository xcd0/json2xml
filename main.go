package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/pkg/errors"
)

var (
	version  string = "debug build" // makefileからビルドされると上書きされる。
	revision string

	Revision = func() string { // {{{
		revision := ""
		modified := false
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					//return setting.Value
					revision = setting.Value
					if len(setting.Value) > 7 {
						revision = setting.Value[:7] // 最初の7文字にする
					}
				}
				if setting.Key == "vcs.modified" {
					modified = setting.Value == "true"
				}
			}
		}
		if modified {
			revision = "develop+" + revision
		}
		return revision
	}() // }}}
	parser *arg.Parser // ShowHelp() で使う

	toJsonFromXml = false // xmlからjsonに変換する。
)

func main() {

	log.SetFlags(log.Ltime | log.Lshortfile) // ログの出力書式を設定する
	if len(os.Args) == 1 {
		// 標準入力から読み取り、標準出力に出力する。
		ret := XmlJsonConverter(os.Stdin, toJsonFromXml)
		fmt.Println(ret)
		os.Exit(0)
	}
	args := ArgParse()

	for _, path := range args.Input {
		r, err := os.Open(path)
		if err != nil {
			panic(errors.Errorf("%v", err))
		}
		defer r.Close()

		ret := XmlJsonConverter(r, toJsonFromXml)

		if args.Debug {
			log.Println(ret)
		}
		out := replaceExt(path, ".json")
		log.Printf("%v -> %v", path, out)
		os.RemoveAll(out)
		WriteText(out, ret)
	}
}

/*
	func GetText(path string) string {
		b, err := os.ReadFile(path) // https://pkg.go.dev/os@go1.20.5#ReadFile
		if err != nil {
			panic(errors.Errorf("Error: %v, file: %v", err, path))
		}
		str := string(b)
		return str
	}

	func GetFilePathWithoutExt(path string) string {
		return filepath.ToSlash(filepath.Join(filepath.Dir(path), GetFileNameWithoutExt(path)))
	}
*/
func GetFileNameWithoutExt(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}

func WriteText(file, str string) {
	f, err := os.Create(file)
	defer f.Close()
	if err != nil {
		panic(errors.Errorf("%v", err))
	} else {
		if _, err := f.Write([]byte(str)); err != nil {
			panic(errors.Errorf("%v", err))
		}
	}
}
func replaceExt(filePath, to string) string {
	ext := filepath.Ext(filePath)
	if len(to) <= 0 {
		return filePath
	}
	return filePath[:len(filePath)-len(ext)] + to
}

type Args struct { //{{{
	//Input      string       `arg:"-i,--input"         help:"入力ファイル。"`
	Input   []string `arg:"positional"         help:"入力ファイル。"`
	Debug   bool     `arg:"-d,--debug"         help:"デバッグ用。ログが詳細になる。"`
	Version bool     `arg:"-v,--version"       help:"バージョン情報を出力する。"`
	//VersionSub *ArgsVersion `arg:"subcommand:version" help:"バージョン情報を出力する。"`
}
type ArgsVersion struct {
}

func (args *Args) Print() {
	//	log.Printf(`
	//
	// Csv  : %v
	// Row  : %v
	// Col  : %v
	// Grep : %v
	// `, args.Csv, args.Row, args.Col, args.Grep)
}

//}}}

func ShowHelp(post string) { // {{{
	buf := new(bytes.Buffer)
	parser.WriteHelp(buf)
	fmt.Printf("%v\n", strings.ReplaceAll(buf.String(), "display this help and exit", "ヘルプを出力する。"))
	if len(post) != 0 {
		fmt.Println(post)
	}
	os.Exit(1)
}
func ShowVersion() {
	if len(Revision) == 0 {
		// go installでビルドされた場合、gitの情報がなくなる。その場合v0.0.0.のように末尾に.がついてしまうのを避ける。
		fmt.Printf("%v version %v\n", GetFileNameWithoutExt(os.Args[0]), version)
	} else {
		fmt.Printf("%v version %v.%v\n", GetFileNameWithoutExt(os.Args[0]), version, revision)
	}
	os.Exit(0)
} // }}}

func ArgParse() *Args {
	args := &Args{}
	var err error
	parser, err = arg.NewParser(arg.Config{Program: GetFileNameWithoutExt(os.Args[0]), IgnoreEnv: false}, args)
	if err != nil {
		ShowHelp(fmt.Sprintf("%v", errors.Errorf("%v", err)))
	}
	if err := parser.Parse(os.Args[1:]); err != nil {
		if err.Error() == "help requested by user" {
			ShowHelp("")
		} else if err.Error() == "version requested by user" {
			ShowVersion()
		} else {
			panic(errors.Errorf("%v", err))
		}
	}
	if args.Version {
		ShowVersion()
	}
	if args.Debug {
		args.Print()
	}
	return args
}
