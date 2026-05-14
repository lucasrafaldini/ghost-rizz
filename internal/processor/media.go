package processor

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/dsoprea/go-exif/v3"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
	pngstructure "github.com/dsoprea/go-png-image-structure/v2"
)

// MediaHandler defines an interface for parsing and modifying EXIF metadata across different image formats.
type MediaHandler interface {
	DropExif() error
	SetExif(ib *exif.IfdBuilder) error
	Write(w io.Writer) error
	RawExif() ([]byte, error)
}

// GetMediaHandler returns a MediaHandler for the given file path based on its extension.
func GetMediaHandler(path string) (MediaHandler, error) {
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") {
		jmp := jpegstructure.NewJpegMediaParser()
		intfc, err := jmp.ParseFile(path)
		if err != nil {
			return nil, err
		}
		return &jpegHandler{sl: intfc.(*jpegstructure.SegmentList)}, nil
	} else if strings.HasSuffix(lower, ".png") {
		pmp := pngstructure.NewPngMediaParser()
		intfc, err := pmp.ParseFile(path)
		if err != nil {
			return nil, err
		}
		return &pngHandler{cs: intfc.(*pngstructure.ChunkSlice)}, nil
	} else if strings.HasSuffix(lower, ".heic") || strings.HasSuffix(lower, ".heif") {
		return &heicHandler{filepath: path}, nil
	}
	return nil, fmt.Errorf("unsupported file format: %s", path)
}

// --- Dependency Checking ---

// CheckExifTool verifies if exiftool is available in the system PATH.
func CheckExifTool() error {
	if _, err := exec.LookPath("exiftool"); err != nil {
		return fmt.Errorf("exiftool is required for HEIC processing but was not found in PATH")
	}
	return nil
}

// --- JPEG Handler ---

type jpegHandler struct {
	sl *jpegstructure.SegmentList
}

func (h *jpegHandler) DropExif() error {
	_, err := h.sl.DropExif()
	return err
}

func (h *jpegHandler) SetExif(ib *exif.IfdBuilder) error {
	return h.sl.SetExif(ib)
}

func (h *jpegHandler) Write(w io.Writer) error {
	return h.sl.Write(w)
}

func (h *jpegHandler) RawExif() ([]byte, error) {
	_, data, err := h.sl.Exif()
	return data, err
}

// --- PNG Handler ---

type pngHandler struct {
	cs *pngstructure.ChunkSlice
}

func (h *pngHandler) DropExif() error {
	exifChunk, err := h.cs.FindExif()
	if err != nil {
		if errors.Is(err, exif.ErrNoExif) || strings.Contains(strings.ToLower(err.Error()), "no exif") {
			return nil
		}
		return err
	}
	if exifChunk != nil {
		chunks := h.cs.Chunks()
		newChunks := make([]*pngstructure.Chunk, 0, len(chunks)-1)
		for _, c := range chunks {
			if c != exifChunk {
				newChunks = append(newChunks, c)
			}
		}
		h.cs = pngstructure.NewChunkSlice(newChunks)
	}
	return nil
}

func (h *pngHandler) SetExif(ib *exif.IfdBuilder) error {
	return h.cs.SetExif(ib)
}

func (h *pngHandler) Write(w io.Writer) error {
	return h.cs.WriteTo(w)
}

func (h *pngHandler) RawExif() ([]byte, error) {
	_, data, err := h.cs.Exif()
	return data, err
}

// --- HEIC Handler ---

type heicHandler struct {
	filepath string
	mode     string
}

func (h *heicHandler) DropExif() error {
	h.mode = "clean"
	return nil
}

// SetExif for HEIC files is currently unsupported.
// Unlike JPEG/PNG handlers, this implementation cannot apply the supplied
// IfdBuilder to the output file, so it must fail explicitly rather than
// silently ignoring the requested metadata changes.
func (h *heicHandler) SetExif(ib *exif.IfdBuilder) error {
	h.mode = "fuzz"
	if ib != nil {
		return errors.New("setting EXIF from IfdBuilder is unsupported for HEIC")
	}
	return nil
}

func (h *heicHandler) Write(w io.Writer) error {
	if h.mode == "" {
		input, err := os.ReadFile(h.filepath)
		if err != nil {
			return err
		}
		_, err = w.Write(input)
		return err
	}

	tmpFile, err := os.CreateTemp("", "ghost-rizz-*.heic")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer func() { _ = tmpFile.Close(); _ = os.Remove(tmpName) }()

	src, err := os.Open(h.filepath)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	if _, err = io.Copy(tmpFile, src); err != nil {
		return err
	}
	_ = tmpFile.Close() // Close so exiftool can work on it

	var cmd *exec.Cmd
	switch h.mode {
	case "clean":
		cmd = exec.Command("exiftool", "-all=", "-overwrite_original", "--", tmpName)
	case "fuzz":
		cmd = exec.Command("exiftool", "-all=", "-overwrite_original", "-Make="+generateRandomString(10), "-Model="+generateRandomString(12), "-Software="+generateRandomString(15), "--", tmpName)
	}

	if cmd != nil {
		out, err := cmd.CombinedOutput()
		if err != nil || strings.Contains(string(out), "0 image files updated") {
			return fmt.Errorf("exiftool error: %v, output: %s", err, string(out))
		}
	}

	updated, err := os.Open(tmpName)
	if err != nil {
		return err
	}
	defer func() { _ = updated.Close() }()

	_, err = io.Copy(w, updated)
	return err
}

func (h *heicHandler) RawExif() ([]byte, error) {
	cmd := exec.Command("exiftool", "-b", "-exif", h.filepath)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no exif data")
	}
	return out, nil
}
