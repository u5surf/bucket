package odatas

import (
	"log"
	"reflect"

	"github.com/couchbase/gocb"
)

const (
	//BucketUsername = "Administrator"
	//BucketPassword = "password"
	TagJson       = "json"
	TagIndexable  = "indexable"
	TagReferenced = "referenced"
)

type indexer struct {
	bucket        *gocb.Bucket
	bucketManager *gocb.BucketManager
}

func NewIndexer(b *gocb.Bucket, BucketUsername, BucketPassword string) *indexer {
	return &indexer{bucket: b, bucketManager: b.Manager(BucketUsername, BucketPassword)}
}

func (i *indexer) Index(v interface{}) error {
	if err := i.bucketManager.CreatePrimaryIndex("", true, false); err != nil {
		log.Fatalf("Error when create primary index %+v", err)
	}

	t := reflect.TypeOf(v)

	indexable, references := goDeep(t)

	if err := makeIndex(i.bucketManager, t.Name(), indexable); err != nil {
		return err
	}

	if err := makeReference(i.bucketManager, v, references); err != nil {
		return err
	}

	return nil
}

func goDeep(t reflect.Type) (indexed []string, referenced []string) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type.Kind() == reflect.Struct {
			goDeep(f.Type)
		}
		if f.Tag != "" {
			if json := f.Tag.Get(TagJson); json != "" && json != "-" {
				if f.Tag.Get(TagIndexable) != "" {
					indexed = append(indexed, json)
				}
				if f.Tag.Get(TagReferenced) != "" {
					referenced = append(referenced, json)
				}
			}
		}
	}
	return indexed, referenced
}

func makeIndex(manager *gocb.BucketManager, indexName string, indexedFields []string) error {
	if err := manager.CreateIndex(indexName, indexedFields, false, false); err != nil {
		if err == gocb.ErrIndexAlreadyExists {
			_ = manager.DropIndex(indexName, true)
			return makeIndex(manager, indexName, indexedFields)
		} else {
			log.Printf("Error when create secondary index %+v", err)
			return err
		}
	}
	return nil
}

func makeReference(bm *gocb.BucketManager, v interface{}, referenced []string) error {
	return nil
}
