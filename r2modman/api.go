package r2modman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type APIPackageResponse struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Versions []struct {
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		FileSize    int64  `json:"file_size"`
		DownloadURL string `json:"download_url"`
	} `json:"versions"`
}

// FindVersion returns the download url and file size of the given full version name
// (e.g. "Azumatt-AAA_Crafting-2.1.10"), and whether that version exists in this package.
func (p *APIPackageResponse) FindVersion(fullName string) (downloadURL string, fileSize int64, ok bool) {
	for _, v := range p.Versions {
		if v.FullName == fullName {
			return v.DownloadURL, v.FileSize, true
		}
	}
	return "", 0, false
}

var myClient = &http.Client{}

func GetPackagesMetadata(ctx context.Context, metadataUrl string) (packages map[string]*APIPackageResponse, err error) {

	req, err := http.NewRequestWithContext(ctx, "GET", metadataUrl, nil)
	if err != nil {
		return
	}

	r, err := myClient.Do(req)
	if err != nil {
		return
	}
	defer r.Body.Close()

	if r.StatusCode > 299 {
		err = fmt.Errorf("unable to fetch %s, HTTP Error %d", metadataUrl, r.StatusCode)
		return
	}

	var packageMetadataList []*APIPackageResponse
	err = json.NewDecoder(r.Body).Decode(&packageMetadataList)
	if err != nil {
		return
	}

	packages = map[string]*APIPackageResponse{}
	for _, p := range packageMetadataList {
		packages[p.FullName] = p
	}
	return
}
