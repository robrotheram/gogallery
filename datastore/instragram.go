package datastore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ahmdrz/goinsta/v2"
	"github.com/dgraph-io/badger"
	"github.com/disintegration/imaging"
	"github.com/robrotheram/gogallery/worker"
)

type Instagram struct {
	insta       *goinsta.Instagram
	GalleryPath string
	db          *badger.DB
}

func (i *Instagram) SetUpAlbum() {
	worker.MakeThumbnail("themes/default/static/img/instagram.png")
	p := Picture{
		Id:   "themes/default/static/img/instagram.png",
		Name: "instagram",
		Path: "themes/default/static/img/instagram.png",
	}
	err := Cache.Tables("PICTURE").Save(p)
	fmt.Println(err)
	Cache.Tables("ALBUM").Save(Album{
		Id:         i.GalleryPath + "/instagram",
		Name:       "instagram",
		ProfileIMG: &p,
		ModTime:    time.Now(),
		Parent:     ""})
}

func (i *Instagram) Connect(username, password string) error {
	insta := goinsta.New(username, password)
	if err := insta.Login(); err != nil {
		return err
	}
	i.insta = insta
	i.db = createDatastore("instagram")
	return nil
}

func (i *Instagram) UploadPhoto(path, caption string) error {
	img, err := imaging.Open(path)
	if err != nil {
		return err
	}
	maxSize := img.Bounds().Dy()
	if img.Bounds().Dx() < img.Bounds().Dy() {
		maxSize = img.Bounds().Dx()
	}
	fmt.Printf("width: %d, height: %d, max: %d \n", img.Bounds().Dx(), img.Bounds().Dy(), maxSize)
	//	crop from center
	centercropimg := imaging.CropCenter(img, maxSize, maxSize)
	dstImage128 := imaging.Resize(centercropimg, 1024, 1024, imaging.Lanczos)
	buf := new(bytes.Buffer)
	imaging.Encode(buf, dstImage128, imaging.JPEG)

	post, uploadErr := i.insta.UploadPhoto(buf, caption, 50, 0)
	if uploadErr != nil {
		return uploadErr
	}
	i.savePost(post)

	return nil
}
func (i *Instagram) getImagePathName(item goinsta.Item) string {
	return fmt.Sprintf("%s-%s.jpg", "ig-post", time.Unix(item.TakenAt, 0).Format("02-01-2006 15:04:05"))
}
func (i *Instagram) getImageName(item goinsta.Item) string {
	return fmt.Sprintf("%s-%s", "ig-post", time.Unix(item.TakenAt, 0).Format("02-01-2006 15:04:05"))
}
func (i *Instagram) SyncFrom() error {
	latest := i.insta.Account.Feed()
	for latest.Next(false) {

		for _, item := range latest.Items {
			if item.MediaType == 1 {
				p, e := i.GetPost(item.ID)
				if (e == nil || e == badger.ErrKeyNotFound) && len(p.ID) == 0 {

					_, _, downloadERR := item.Download(i.GalleryPath+"/instagram", i.getImagePathName(item))
					if downloadERR != nil {
						fmt.Printf("Error Downloading from Instagram: %v", downloadERR)
					} else {
						fmt.Printf("Downloading Post ID : %s \n", item.ID)
						i.savePost(item)
						p := Picture{
							Id:         i.GalleryPath + "/instagram/images/" + i.getImagePathName(item),
							Name:       i.getImageName(item),
							Path:       i.GalleryPath + "/instagram/images/" + i.getImagePathName(item),
							Album:      "instagram",
							PostedToIG: true,
							Caption:    item.Caption.Text,
							Exif: Exif{
								Camera:    "Instagram",
								DateTaken: time.Unix(item.TakenAt, 0),
							}}

						Cache.Tables("PICTURE").Save(p)
						worker.MakeThumbnail(i.GalleryPath + "/instagram/images/" + i.getImageName(item))
						worker.MakeLargeThumbnail(i.GalleryPath + "/instagram/images/" + i.getImageName(item))
					}
				} else {
					pics, err := Cache.Tables("PICTURE").Query("Name", i.getImageName(item), 1)
					if err != nil {
						break
					}
					if len(pics.([]Picture)) == 0 {
						break
					}
					picture := pics.([]Picture)[0]
					picture.Caption = item.Caption.Text
					Cache.Tables("PICTURE").Save(picture)
				}
			}

		}
		if err := latest.Error(); err != nil {
			if err := latest.Error(); err == goinsta.ErrNoMore {
				break
			}
		}
	}
	return nil
}

func (i *Instagram) savePost(original goinsta.Item) error {

	err := i.db.Update(func(tx *badger.Txn) error {
		return tx.Set([]byte(original.ID), serialize(original))
	})
	return err

}

func (i *Instagram) GetPost(id string) (goinsta.Item, error) {
	post := goinsta.Item{}
	err := i.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get([]byte(id))
		if err != nil {
			return err
		}
		valCopy, err := item.ValueCopy(nil)
		post, _ = deserialize(valCopy)
		return nil
	})
	if err != nil {
		return goinsta.Item{}, err
	}
	return post, nil
}

func (i *Instagram) GetAllPosts() ([]goinsta.Item, error) {
	posts := []goinsta.Item{}
	err := i.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var data []byte
			err := item.Value(func(v []byte) error {
				data = v
				return nil
			})
			if err != nil {
				return err
			}
			u, error := deserialize(data)
			if error != nil {
				return error
			}
			posts = append(posts, u)
		}
		return nil
	})
	return posts, err
}

func serialize(u goinsta.Item) []byte {
	b, _ := json.Marshal(u)
	return b
}

func deserialize(b []byte) (goinsta.Item, error) {
	var u = goinsta.Item{}
	err := json.Unmarshal(b, &u)
	return u, err
}
