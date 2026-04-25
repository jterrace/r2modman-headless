package r2modman

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type ExportParser interface {
	Parse(file string) (*ExportR2x, error)
}

type exportParserImpl struct {
}

// ParseModString takes a Thunderstore dependency string and converts it to an ExportR2xMod
//
// e.g. incoming "Azumatt-AzuExtendedPlayerInventory-2.4.1"
//
// becomes
//
//	ExportR2xMod{
//		Name: "Azumatt-AzuExtendedPlayerInventory",
//		Version: ExportR2xModVersion{
//			Major: 2,
//			Minor: 4,
//			Patch: 1,
//		},
//		Enabled: true,
//	}
func ParseModString(rawString string) (*ExportR2xMod, error) {
	lastHyphenIdx := strings.LastIndex(rawString, "-")
	if lastHyphenIdx == -1 {
		return nil, fmt.Errorf("invalid mod string format: missing hyphen in '%s'", rawString)
	}

	namePart := rawString[:lastHyphenIdx]
	versionPart := rawString[lastHyphenIdx+1:]

	vParts := strings.Split(versionPart, ".")
	if len(vParts) != 3 {
		return nil, fmt.Errorf("invalid version format: expected X.Y.Z, got '%s'", versionPart)
	}

	major, err := strconv.Atoi(vParts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse major version: %w", err)
	}
	minor, err := strconv.Atoi(vParts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse minor version: %w", err)
	}
	patch, err := strconv.Atoi(vParts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse patch version: %w", err)
	}

	return &ExportR2xMod{
		Name: namePart,
		Version: ExportR2xModVersion{
			Major: major,
			Minor: minor,
			Patch: patch,
		},
		Enabled: true,
	}, nil
}

func (p *exportParserImpl) Parse(file string) (*ExportR2x, error) {

	reader, err := zip.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var exportMetadata *ExportR2x

	for _, v := range reader.File {

		fileread, err := v.Open()
		if err != nil {
			return nil, err
		}
		defer fileread.Close()

		// parse metadata file
		if v.Name == "export.r2x" {
			content, err := io.ReadAll(fileread)
			if err != nil {
				return nil, err
			}
			err = yaml.Unmarshal(content, &exportMetadata)
			if err != nil {
				return nil, err
			}
		}

		if v.Name == "manifest.json" {
			content, err := io.ReadAll(fileread)
			if err != nil {
				return nil, err
			}
			var manifest ModpackManifest
			if err = json.Unmarshal(content, &manifest); err != nil {
				return nil, err
			}

			exportMetadata = &ExportR2x{
				ProfileName: manifest.Name,
			}
			for _, v := range manifest.Dependencies {
				mod, err := ParseModString(v)
				if err != nil {
					return nil, fmt.Errorf("failed to parse dependency '%s': %w", v, err)
				}
				exportMetadata.Mods = append(exportMetadata.Mods, *mod)
			}
		}
	}

	if exportMetadata == nil {
		return nil, fmt.Errorf("%s does not contain an export.r2x file", file)
	}

	return exportMetadata, nil
}

func newExportParser() (ExportParser, error) {
	return &exportParserImpl{}, nil
}
