package datastore

import (
	"github.com/dgraph-io/badger"
	"github.com/robrotheram/gogallery/config"
	"log"
	"os"
	"reflect"
)

type DataStore struct {
	dataFactories map[string]DS
}

type DS interface {
	Close()
	Initialize()
	Get(id string) (interface{}, error)
	Query(field string, val interface{}, limit int) (interface{}, error)
	Edit(obj interface{}) error
	Delete(obj interface{}) error
	GetAll() (interface{}, error)
	Save(obj interface{}) error
}

var Cache *DataStore

func NewDataStore() *DataStore {
	d := DataStore{}
	d.dataFactories = make(map[string]DS)
	d.RegisterData("ALBUM", albumDataStore{}.New())
	d.RegisterData("PICTURE", pictureDataStore{}.New())
	return &d
}

func (d *DataStore) Load() {
	for _, ds := range d.dataFactories {
		ds.Initialize()
	}
}

func (d *DataStore) RegisterData(name string, factory DS) {
	if factory == nil {
		log.Panicf("datastore factory %s does not exist.", name)
	}
	_, registered := d.dataFactories[name]
	if registered {
		log.Printf("datastore factory %s already registered. Ignoring. \n", name)
	}
	d.dataFactories[name] = factory
}

func (d *DataStore) Close() {
	for _, v := range d.dataFactories {
		v.Close()
	}

}

func (d DataStore) DoesTableExist(table string) bool {
	return d.dataFactories[table] != nil
}

func (d DataStore) Tables(table string) DS {
	return d.dataFactories[table]
}

// Helper function
func createDatastore(ds string) *badger.DB {
	opts := badger.DefaultOptions
	log.Println("DB location:" + config.Config.Database.Baseurl)
	opts.Dir = config.Config.Database.Baseurl + ds
	opts.ValueDir = config.Config.Database.Baseurl + ds

	os.MkdirAll(opts.Dir, os.ModePerm)

	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	return db
}

func getFieldByName(str interface{}, field string) interface{} {
	r := reflect.ValueOf(str)
	f := reflect.Indirect(r).FieldByName(field)
	switch f.Type().Name() {
	case "int":
		return f.Int()
	case "string":
		return f.String()
	}
	return nil
}
