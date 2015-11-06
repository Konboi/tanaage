package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/goauth2/oauth/jwt"
	"code.google.com/p/google-api-go-client/oauth2/v2"
	"google.golang.org/api/drive/v2"
)

var (
	uploadedFiles        = map[string]UploadFile{}
	GOOGLE_SERVICE_SCOPE = []string{"https://www.googleapis.com/auth/drive"}
	HISTORY_JSON         = ".history.json"
)

type UploadFile struct {
	Name         string                   `json:"name"`
	LastUpdateAt time.Time                `json:"last_update_at"`
	Folder       []*drive.ParentReference `json:"folder"`
	FileId       string                   `json:"file_id"`
}

type Uploader struct {
	config       *Config
	driveService *drive.Service
}

func NewUploader(config *Config) (*Uploader, error) {
	uploader := &Uploader{
		config: config,
	}

	return uploader, nil
}

func (uploader *Uploader) Check() error {
	for _, upload := range uploader.config.Uploads {
		_, err := os.Stat(upload.From)
		if err != nil {
			return err
		}

	}

	return nil
}

func (uploader *Uploader) Prepare() error {
	err := uploader.initGoogleService()
	if err != nil {
		return err
	}

	err = uploader.checkHistory()
	if err != nil {
		return err
	}

	return nil
}

func (uploader *Uploader) checkHistory() error {
	_, err := os.Stat(HISTORY_JSON)

	if err != nil && strings.Contains(err.Error(), "no such file or directory") {
		return nil
	} else if err != nil {
		return err
	}

	file, err := ioutil.ReadFile(HISTORY_JSON)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &uploadedFiles)
	if err != nil {
		return err
	}

	return nil
}

func (uploader *Uploader) Run() error {
	for _, upload := range uploader.config.Uploads {
		file, err := os.Stat(upload.From)
		if err != nil {
			return err
		}

		if file.IsDir() {
			err = filepath.Walk(upload.From, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					media, err := os.Open(path)

					if err != nil {
						return err
					}

					defer media.Close()

					if uploadedFiles[media.Name()].Name == "" {
						uploadTo := strings.Replace(media.Name(), upload.From, upload.To, 1)
						result, err := uploader.uploadFile(media, uploadTo)
						if err != nil {
							writeJson()
							return err
						}

						uploadedFiles[media.Name()] = UploadFile{
							Name:         result.Title,
							LastUpdateAt: info.ModTime(),
							Folder:       result.Parents,
							FileId:       result.Id,
						}

					} else if uploadedFiles[media.Name()].Name != "" && uploadedFiles[media.Name()].LastUpdateAt != info.ModTime() {
						result, err := uploader.updateFile(media, uploadedFiles[media.Name()])
						if err != nil {
							writeJson()
							return err
						}

						uploadedFiles[media.Name()] = UploadFile{
							Name:         result.Title,
							LastUpdateAt: info.ModTime(),
							Folder:       result.Parents,
							FileId:       result.Id,
						}
					}

				}

				return nil
			})

			if err != nil {
				return err
			}
		} else {
			media, _ := os.Open(upload.From)
			defer media.Close()

			var result *drive.File
			if uploadedFiles[media.Name()].Name == "" {
				uploadTo := strings.Replace(media.Name(), upload.From, upload.To, 1)
				result, err = uploader.uploadFile(media, uploadTo)

				if err != nil {
					writeJson()
					return err
				}

			} else if uploadedFiles[media.Name()].LastUpdateAt != file.ModTime() {
				result, err = uploader.updateFile(media, uploadedFiles[media.Name()])

				if err != nil {
					writeJson()
					return err
				}
			}

			uploadedFiles[media.Name()] = UploadFile{
				Name:         result.Title,
				LastUpdateAt: file.ModTime(),
				Folder:       result.Parents,
				FileId:       result.Id,
			}
		}

		if err != nil {
			return err
		}
	}

	err := writeJson()
	if err != nil {
		return err
	}

	return nil
}

func (uploader *Uploader) initGoogleService() error {
	token := jwt.NewToken(
		uploader.config.ClientEmail,
		strings.Join(GOOGLE_SERVICE_SCOPE, " "),
		[]byte(uploader.config.PrivateKey),
	)

	client := &http.Client{}

	oauthToken, err := token.Assert(client)
	if err != nil {
		return err
	}

	transport := &oauth.Transport{
		Token: oauthToken,
	}

	c := transport.Client()

	_, err = oauth2.New(c)
	if err != nil {
		return err
	}

	driveService, err := drive.New(c)
	if err != nil {
		return err
	}

	uploader.driveService = driveService

	return nil
}

func (uploader *Uploader) createDir(name string, parentId string) (result *drive.File, err error) {
	parent := &drive.ParentReference{
		Id: parentId,
	}

	driveFile := &drive.File{
		Title:    name,
		Parents:  []*drive.ParentReference{parent},
		MimeType: "application/vnd.google-apps.folder",
	}

	result, err = uploader.driveService.Files.Insert(driveFile).Do()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (uploader *Uploader) uploadFile(media *os.File, path string) (result *drive.File, err error) {
	parentFolderId := uploader.config.Folder
	fileTitle := ""

	for _, folder := range strings.Split(strings.Replace(path, "//", "/", -1), "/") {
		list, _ := uploader.driveService.Files.List().Q(fmt.Sprintf("title='%s'", folder)).OrderBy("folder,createdDate").Do()
		if !strings.Contains(folder, ".") {
			items := map[string]string{}
			for _, item := range list.Items {
				items[item.Title] = item.Id
			}

			if len(list.Items) < 1 || items[folder] == "" {
				result, err := uploader.createDir(folder, parentFolderId)
				parentFolderId = result.Id

				if err != nil {
					return nil, err
				}
			} else {
				parentFolderId = items[folder]
			}
		} else {
			fileTitle = folder
		}
	}

	parent := &drive.ParentReference{
		Id: parentFolderId,
	}
	driveFile := &drive.File{
		Title:   fileTitle,
		Parents: []*drive.ParentReference{parent},
	}

	result, err = uploader.driveService.Files.Insert(driveFile).Media(media).Do()

	if err != nil {
		return nil, err
	}

	log.Println(strings.Replace(path, "//", "/", -1), "uploaded")

	return result, nil
}

func (uploader *Uploader) updateFile(media *os.File, file UploadFile) (result *drive.File, err error) {
	driveFile := &drive.File{
		Title: file.Name,
	}

	result, err = uploader.driveService.Files.Update(file.FileId, driveFile).Media(media).Do()
	if err != nil {
		return nil, err
	}

	log.Println(file.Name, " updated")

	return result, nil

}

func writeJson() error {
	j, err := json.MarshalIndent(uploadedFiles, "", "  ")
	if err != nil {
		return err
	}

	ioutil.WriteFile(HISTORY_JSON, j, os.ModePerm)

	return nil
}
