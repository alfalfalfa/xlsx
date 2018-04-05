package xlsx

import (
	"reflect"

	"fmt"

	"os"
	"sync"

	"crypto/sha256"
	"io/ioutil"
	"path/filepath"

	"github.com/ugorji/go/codec"
)

var CacheEnable bool
var CacheStateDump bool

const BOOK_CACHE_FORMAT = "cache_xlsx_book_%d_%x"
const SHEET_CACHE_FORMAT = "cache_xlsx_sheet_%s_%x"

var mh codec.MsgpackHandle

//var dec *codec.Decoder
//var enc *codec.Encoder
var encPool sync.Pool
var decPool sync.Pool

func init() {
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mh.StructToArray = true
	b := make([]byte, 0)
	decPool = sync.Pool{
		New: func() interface{} { return codec.NewDecoderBytes(nil, &mh) },
	}
	encPool = sync.Pool{
		New: func() interface{} { return codec.NewEncoderBytes(&b, &mh) },
	}
}

func getEnc() *codec.Encoder {
	return encPool.Get().(*codec.Encoder)
}
func putEnc(v *codec.Encoder) {
	b := make([]byte, 0)
	v.ResetBytes(&b)
	encPool.Put(v)
}
func getDec() *codec.Decoder {
	return decPool.Get().(*codec.Decoder)
}
func putDec(v *codec.Decoder) {
	v.ResetBytes(nil)
	decPool.Put(v)
}
func nameForSheet(sheet xlsxSheet, sheetXMLMap map[string]string) string {
	sheetName, ok := sheetXMLMap[sheet.Id]
	if !ok {
		if sheet.SheetId != "" {
			return fmt.Sprintf("sheet%s", sheet.SheetId)
		} else {
			return fmt.Sprintf("sheet%s", sheet.Id)
		}
	}
	return sheetName
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func LoadBookCache(path string, v interface{}) bool {
	if !CacheEnable {
		return false
	}
	s, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	cachePath := filepath.Join(os.TempDir(), fmt.Sprintf(BOOK_CACHE_FORMAT, s.ModTime().Unix(), sha256.Sum256([]byte(path))))
	if !exists(cachePath) {
		return false
	}
	cache, err := ioutil.ReadFile(cachePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}
	dec := getDec()
	defer putDec(dec)
	dec.ResetBytes(cache)
	err = dec.Decode(v)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}
	return true
}

func SaveBookCache(path string, v interface{}) bool {
	if !CacheEnable {
		return false
	}
	s, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	cachePath := filepath.Join(os.TempDir(), fmt.Sprintf(BOOK_CACHE_FORMAT, s.ModTime().Unix(), sha256.Sum256([]byte(path))))
	cache := make([]byte, 0, 1024*8)
	enc := getEnc()
	defer putEnc(enc)
	enc.ResetBytes(&cache)
	err = enc.Encode(v)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	if CacheStateDump{
		fmt.Println("save cache", cachePath, len(cache))
	}

	err = ioutil.WriteFile(cachePath, cache, os.ModePerm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}
	return true
}

func LoadSheetCache(sheetName string, sheetData []byte, v **xlsxWorksheet) bool {
	if !CacheEnable {
		return false
	}
	cachePath := filepath.Join(os.TempDir(), fmt.Sprintf(SHEET_CACHE_FORMAT, sheetName, sha256.Sum256(sheetData)))
	if !exists(cachePath) {
		return false
	}
	cache, err := ioutil.ReadFile(cachePath)
	if err != nil {
		//	fmt.Fprintln(os.Stderr, err.Error())
		//	return false
	}
	dec := getDec()
	defer putDec(dec)
	dec.ResetBytes(cache)
	err = dec.Decode(v)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	if CacheStateDump{
		fmt.Println("load cache", cachePath)
	}

	return true
}

func SaveSheetCache(sheetName string, sheetData []byte, v *xlsxWorksheet) bool {
	if !CacheEnable {
		return false
	}
	cachePath := filepath.Join(os.TempDir(), fmt.Sprintf(SHEET_CACHE_FORMAT, sheetName, sha256.Sum256(sheetData)))

	cache := make([]byte, 0, 1024*8)
	enc := getEnc()
	defer putEnc(enc)
	enc.ResetBytes(&cache)
	err := enc.Encode(v)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	if CacheStateDump{
		fmt.Println("save cache", cachePath, len(cache))
	}

	err = ioutil.WriteFile(cachePath, cache, os.ModePerm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}
	return true
}
