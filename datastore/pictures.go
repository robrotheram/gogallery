package datastore

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/nfnt/resize"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"github.com/rwcarlsen/goexif/tiff"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Exif struct {
	FStop        float64 `json:"f_stop"`
	FocalLength  float64 `json:"focal_length"`
	ShutterSpeed string  `json:"shutter_speed"`
	ISO          string  `json:"iso"`
	Dimension    string  `json:"dimension"`
	Camera       string  `json:"camera"`
	LensModel    string  `json: lens_model`
}

type Picture struct {
	Id         string    `json:"id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	ModTime    time.Time `json:"mod_time"`
	FormatTime string    `json:"format_time"`
	Album      string    `json:"album"`
	Exif       Exif      `json:"exif"`
}

func (u *Picture) serialize() []byte {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	if err := e.Encode(u); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func (u *Picture) deserialize(b []byte) error {
	dCache := bytes.NewBuffer(b)
	d := gob.NewDecoder(dCache)
	if err := d.Decode(u); err != nil {
		return err
	}
	return nil
}

func fnumber(f string) float64 {
	f = strings.Replace(f, "\"", "", -1)
	calc := strings.Split(f, "/")
	if len(calc) != 2 {
		return 10
	}
	if i, err := strconv.ParseFloat(calc[0], 64); err == nil {
		if j, err := strconv.ParseFloat(calc[1], 64); err == nil {
			return i / j
		}
	}
	return 11
}

func decodeExifTag(exf *exif.Exif, tag exif.FieldName) (val string) {
	res, err := exf.Get(tag)
	if err != nil {
		return ""
	}

	switch res.Format() {
	case tiff.StringVal:
		resStr, err := res.StringVal()
		if err != nil {
			fmt.Println(err)
		}
		return resStr
		break
	case tiff.RatVal:
		return strings.Replace(res.String(), "\"", "", -1)
	}
	return res.String()

}

func (u *Picture) CreateExif() {
	f, err := os.Open(u.Path)
	if err == nil {
		exif.RegisterParsers(mknote.All...)
		x, err := exif.Decode(f)
		if err == nil {
			u.Exif = Exif{
				fnumber(decodeExifTag(x, exif.FNumber)),
				fnumber(decodeExifTag(x, exif.FocalLength)),
				decodeExifTag(x, exif.ExposureTime),
				decodeExifTag(x, exif.ISOSpeedRatings),
				fmt.Sprintf("%sx%s", decodeExifTag(x, exif.PixelXDimension), decodeExifTag(x, exif.PixelYDimension)),
				decodeExifTag(x, exif.Make),
				decodeExifTag(x, exif.LensModel)}
		}
	}
}

type pictureDataStore struct {
	db       *badger.DB
	pictures []Picture
}

func (uDs *pictureDataStore) Close() {
	uDs.db.Close()
}
func (u pictureDataStore) New() *pictureDataStore {
	u = pictureDataStore{}
	u.Initialize()
	return &u
}

func (u *pictureDataStore) Initialize() {
	u.db = createDatastore("pictures")
}

func (uDs *pictureDataStore) Save(u interface{}) error {
	original, ok := u.(Picture)
	if ok {
		err := uDs.db.Update(func(tx *badger.Txn) error {
			//fmt.Println(original.Id)
			return tx.Set([]byte(original.Id), original.serialize())
		})
		//fmt.Println(err)
		return err
	}
	fmt.Println("NO ERROR")
	return nil
}

func (uDs *pictureDataStore) Edit(u interface{}) error {
	original, ok := u.(Picture)
	if ok {
		err := uDs.db.Update(func(tx *badger.Txn) error {
			fmt.Println(original.Id)
			return tx.Set([]byte(original.Id), original.serialize())
		})
		fmt.Println(err)
		return err
	}
	fmt.Println("NO ERROR")
	return nil
}

func (uDs *pictureDataStore) Delete(u interface{}) error {
	original, ok := u.(Picture)
	if ok {
		err := uDs.db.Update(func(tx *badger.Txn) error {
			return tx.Delete([]byte(original.Id))
		})
		return err
	}
	return nil
}

func (uDs *pictureDataStore) Get(id string) (interface{}, error) {
	var valCopy []byte
	err := uDs.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get([]byte(id))
		if err != nil {
			return err
		}
		valCopy, err = item.ValueCopy(nil)
		return nil
	})
	if err != nil {
		return Picture{}, err
	}
	u := Picture{}
	u.deserialize(valCopy)
	return u, nil
}

func (uDs *pictureDataStore) GetAll() (interface{}, error) {
	pictures := []Picture{}
	err := uDs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			data, err := item.Value()
			if err != nil {
				return err
			}
			u := Picture{}
			error := u.deserialize(data)
			if error != nil {
				return error
			}
			pictures = append(pictures, u)
		}
		return nil
	})
	return pictures, err
}

func (uDs *pictureDataStore) Query(field string, val interface{}, limit int) (interface{}, error) {
	pictures := []Picture{}
	err := uDs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			if (len(pictures) >= limit) && (limit != 0) {
				return nil
			}
			item := it.Item()
			data, err := item.Value()
			if err != nil {
				return err
			}
			u := Picture{}
			error := u.deserialize(data)
			if error != nil {
				return error
			}
			if getFieldByName(u, field) == val {
				pictures = append(pictures, u)
			}
		}
		return nil
	})
	return pictures, err
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func MakeThumbnail(path string) {
	cachePath := fmt.Sprintf("cache/%s.jpg", GetMD5Hash(path))

	if _, err := os.Stat(cachePath); err == nil {
		return
	}
	fmt.Println("Creating Thumbnail")
	os.MkdirAll("cache", os.ModePerm)
	file, err := os.Open(path)
	if err != nil {
		//fmt.Println(path)
		return
	}
	// decode jpeg into image.Image
	extension := filepath.Ext(path)
	var img image.Image
	var img_err error
	switch extension {
	case ".jpg":
		img, img_err = jpeg.Decode(file)
	case ".png":
		img, img_err = png.Decode(file)
	}
	if img_err != nil {
		fmt.Println(path)
		return
	}
	file.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(300, 0, img, resize.Bilinear)
	out, err := os.Create(cachePath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	jpeg.Encode(out, m, nil)
}
