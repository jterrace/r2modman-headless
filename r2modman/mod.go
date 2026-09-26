package r2modman

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type ModUtil interface {
	// Download fetches the mod zip from downloadURL into the work directory, unless a copy
	// of the expected fileSize is already present.
	Download(mod ExportR2xMod, downloadURL string, fileSize int64) error
}

type modUtilImpl struct {
	config Config
}

func (m *modUtilImpl) verifyIntegrity(mod ExportR2xMod, fileSize int64) error {
	downloadedZipPath := m.config.WorkDirectory + "/" + mod.Filename()
	existingModFile, err := os.Stat(downloadedZipPath)
	if err != nil {
		return fmt.Errorf("mod package %s did not exist", downloadedZipPath)
	}
	log.Printf("Verifying integrity: %s", downloadedZipPath)
	if fileSize == existingModFile.Size() {
		return nil
	}
	log.Printf("File Integrity Failed: %s, expected size: %d bytes, actual size: %d bytes", mod.Filename(), fileSize, existingModFile.Size())

	//package did not validate, delete it
	log.Printf("Removing invalid file: %s", downloadedZipPath)
	err = os.Remove(downloadedZipPath)
	if err != nil {
		return err
	}

	return fmt.Errorf("mod package %s did not validate", downloadedZipPath)
}

func (m *modUtilImpl) Download(mod ExportR2xMod, downloadURL string, fileSize int64) error {
	downloadedZipPath := m.config.WorkDirectory + "/" + mod.Filename()

	client := http.Client{
		Timeout: m.config.ThunderstoreCDNTimeout,
	}

	err := m.verifyIntegrity(mod, fileSize)
	if err == nil && !m.config.ThunderstoreForceDownload {
		return nil // file exists and validates
	}

	log.Printf("Downloading mod: %s from %s", downloadedZipPath, downloadURL)
	resp, err := client.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return fmt.Errorf("unable to download, HTTP Error %d", resp.StatusCode)
	}

	// Create the file
	out, err := os.Create(downloadedZipPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return m.verifyIntegrity(mod, fileSize)
}

func newModUtil(
	config Config,
) (ModUtil, error) {
	return &modUtilImpl{
		config: config,
	}, nil
}
