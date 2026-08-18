package compatibility

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/domain/collection"
	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/companion/internal/ports"
)

// File names are shared with the legacy Python exporter.
const (
	JSONName = "mtga_collection.json"
	CSVName  = "mtga_collection.csv"
	TextName = "mtga_collection.txt"
)

var exportNames = map[ports.ExportFormat]string{
	ports.ExportJSON: JSONName,
	ports.ExportCSV:  CSVName,
	ports.ExportText: TextName,
}

// Writer atomically replaces each compatibility projection in one folder.
type Writer struct {
	Exporter ports.Exporter
}

// WriteAll renders every payload before changing the destination folder.
func (w Writer) WriteAll(ctx context.Context, directory string,
	snapshot collection.Snapshot) (map[ports.ExportFormat]string, error) {
	payloads := make(map[ports.ExportFormat][]byte, len(exportNames))
	for _, format := range []ports.ExportFormat{
		ports.ExportJSON, ports.ExportCSV, ports.ExportText,
	} {
		payload, err := w.Exporter.Export(ctx, snapshot, format)
		if err != nil {
			return nil, err
		}
		payloads[format] = payload
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, fmt.Errorf("compatibility: create output directory: %w", err)
	}
	paths := make(map[ports.ExportFormat]string, len(exportNames))
	for _, format := range []ports.ExportFormat{
		ports.ExportJSON, ports.ExportCSV, ports.ExportText,
	} {
		path := filepath.Join(directory, exportNames[format])
		if err := replaceFile(path, payloads[format]); err != nil {
			return nil, err
		}
		paths[format] = path
	}
	return paths, nil
}

func replaceFile(path string, payload []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".mtga-export-*")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer func() { _ = os.Remove(temporary) }()
	if _, err = file.Write(payload); err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("compatibility: stage %s: %w", filepath.Base(path), err)
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("compatibility: replace %s: %w", filepath.Base(path), err)
	}
	return nil
}
