package service

import (
	"database/sql"
	"io"
	"net/http"
	"os"
	"time"

	"bitbucket.org/pqstudio/go-file/datastore"
	"bitbucket.org/pqstudio/go-file/model"

	"bitbucket.org/pqstudio/go-webutils/slice"

	"bitbucket.org/pqstudio/go-webutils/web"

	"bitbucket.org/pqstudio/go-webutils"
	s3 "bitbucket.org/pqstudio/go-webutils/s3"
)

func Get(tx *sql.Tx, uid string) (*model.File, error) {
	m, err := datastore.Get(tx, uid)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func GetOlderThan(tx *sql.Tx, t int) ([]model.File, error) {
	rs, err := datastore.GetOlderThan(tx, t)
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func Create(tx *sql.Tx, m *model.File) error {
	m.UID = utils.NewUUID()
	m.CreatedAt = time.Now().UTC()

	err := datastore.Create(tx, m)
	if err != nil {
		return err
	}

	return nil
}

func Update(tx *sql.Tx, m *model.File) error {
	err := datastore.Update(tx, m)
	if err != nil {
		return err
	}

	return nil
}

func Delete(tx *sql.Tx, uid string, filename string) error {
	if _, err := os.Stat(filename); err == nil {
		err := os.Remove("/data/tmp/" + filename)
		if err != nil {
			return err
		}
	}

	err := datastore.Delete(tx, uid)
	if err != nil {
		return err
	}

	return nil
}

func DeleteOlderThan(tx *sql.Tx, t int) error {
	rs, err := GetOlderThan(tx, t)

	if err != nil {
		return err
	}

	for _, f := range rs {
		err = Delete(tx, f.UID, f.UniqueID+"-"+f.Name)

		if err != nil {
			continue
		}
	}

	return err
}

func Upload(tx *sql.Tx, fileType string, tmpPath string, uploadData map[string]map[string]interface{}, w http.ResponseWriter, r *http.Request) (*model.File, error) {
	f := &model.File{}

	if len(fileType) == 0 {
		return nil, &web.ValidationError{
			Errors: map[string]string{"error": "no_type"},
		}
	}

	if _, ok := uploadData[fileType]; !ok {
		return nil, &web.ValidationError{
			Errors: map[string]string{"error": "wrong_type"},
		}
	}

	t := uploadData[fileType]
	mime := t["mime"].([]string)

	MaxFileSize := t["size"].(int)

	if r.ContentLength > int64(MaxFileSize) {
		return nil, &web.ValidationError{
			Errors: map[string]string{"error": "file_too_large_cl"},
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(MaxFileSize))
	err := r.ParseMultipartForm(10000000)
	if err != nil {
		return nil, &web.ValidationError{
			Errors: map[string]string{"error": "file_too_large"},
		}
	}

	file, header, err := r.FormFile("file")
	if file != nil {
		defer file.Close()
	}

	if err != nil {
		return nil, &web.ValidationError{
			Errors: map[string]string{"error": "file_not_provided"},
		}
	}

	f.Tmp = true
	f.Type = fileType
	f.Name = header.Filename
	f.UniqueID = utils.NewUUID()
	f.Mime = header.Header.Get("Content-Type")

	if !slice.StringInSlice(f.Mime, mime) {
		return nil, &web.ValidationError{
			Errors: map[string]string{"error": "mime_type_not_allowed"},
		}
	}

	// create directories
	err = os.MkdirAll(tmpPath, 0777)
	if err != nil {
		return nil, err
	}

	err = web.CreateFile(file, tmpPath, f.UniqueID+"-"+f.Name)
	if err != nil {
		return nil, err
	}

	err = Create(tx, f)
	if err != nil {
		return nil, err
	}

	err = DeleteOlderThan(tx, 5)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func GetFromS3(AWSS3Region string, AWSS3Bucket string, filePath string) (io.ReadCloser, error) {
	body, err := s3.Get(AWSS3Region, AWSS3Bucket, filePath)
	defer body.Close()
	return body, err
}

func UploadToS3(AWSS3Region string, AWSS3Bucket string, localFilePath string, filePath string) error {
	fileOpen, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer fileOpen.Close()

	err = s3.Upload(AWSS3Region, AWSS3Bucket, filePath, fileOpen)
	if err != nil {
		return err
	}
	return nil
}

func DeleteFromS3(AWSS3Region string, AWSS3Bucket string, filePath string) error {
	err := s3.Delete(AWSS3Region, AWSS3Bucket, filePath)
	if err != nil {
		return err
	}
	return nil
}
