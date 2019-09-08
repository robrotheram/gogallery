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
}

func (i *Instagram) SetUpAlbum() {
	worker.MakeThumbnail("themes/default/static/img/instagram.png")
	p := Picture{
		Id:   "themes/default/static/img/instagram.png",
		Name: "instagram",
		Path: "themes/default/static/img/instagram.png",
	}
	err := Cache.DB.Save(&p)
	fmt.Println(err)
	Cache.DB.Save(&Album{
		Id:         i.GalleryPath + "/instagram",
		Name:       "instagram",
		ProfileIMG: &p,
		ModTime:    time.Now(),
		Parent:     ""})
}

func (i *Instagram) Connect(username, password string) error {
	insta := goinsta.New(username, password)
	if err := insta.Login(); err != nil {
		fmt.Println(err)
		return err
	}
	i.insta = insta
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
	i.SavePost(post)

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
						i.SavePost(item)
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

						Cache.DB.Save(&p)
						worker.MakeThumbnail(i.GalleryPath + "/instagram/images/" + i.getImageName(item))

					}
				} else {
					Cache.DB.UpdateField(&Picture{Name: i.getImageName(item)}, "Caption", item.Caption.Text)
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

func (i *Instagram) SavePost(original goinsta.Item) error {
	return Cache.DB.Save(original)
}

func (i *Instagram) GetPost(id string) (goinsta.Item, error) {
	var post goinsta.Item
	err := Cache.DB.One("ID", id, &post)
	return post, err
}

func (i *Instagram) GetAllPosts() ([]goinsta.Item, error) {
	var posts []goinsta.Item
	err := Cache.DB.All(&posts)
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
