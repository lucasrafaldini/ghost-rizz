package processor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/dsoprea/go-exif/v3"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
	pngstructure "github.com/dsoprea/go-png-image-structure/v2"
)

type MediaHandler interface {
	DropExif() error
	SetExif(ib *exif.IfdBuilder) error
	Write(w io.Writer) error
	RawExif() ([]byte, error)
}

func GetMediaHandler(filepath string) (MediaHandler, error) {
	ext := strings.ToLower(filepath)
	if strings.HasSuffix(ext, ".jpg") || strings.HasSuffix(ext, ".jpeg") {
		jmp := jpegstructure.NewJpegMediaParser()
		intfc, err := jmp.ParseFile(filepath)
		if err != nil {
			return nil, err
		}
		return &jpegHandler{sl: intfc.(*jpegstructure.SegmentList)}, nil
	} else if strings.HasSuffix(ext, ".png") {
		pmp := pngstructure.NewPngMediaParser()
		intfc, err := pmp.ParseFile(filepath)
		if err != nil {
			return nil, err
		}
		return &pngHandler{cs: intfc.(*pngstructure.ChunkSlice)}, nil
	} else if strings.HasSuffix(ext, ".heic") || strings.HasSuffix(ext, ".heif") {
		return &heicHandler{filepath: filepath}, nil
	}
	return nil, fmt.Errorf("unsupported file format: %s", filepath)
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
	if err == nil && exifChunk != nil {
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

func (h *heicHandler) SetExif(ib *exif.IfdBuilder) error {
	h.mode = "fuzz"
	return nil
}

func (h *heicHandler) Write(w io.Writer) error {
	tmpFile, err := os.CreateTemp("", "ghost-rizz-*.heic")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpName) }()

	input, err := os.ReadFile(h.filepath)
	if err != nil {
		return err
	}
	_ = os.WriteFile(tmpName, input, 0644)

	var cmd *exec.Cmd
	switch h.mode {
	case "clean":
		cmd = exec.Command("exiftool", "-all=", "-overwrite_original", tmpName)
	case "fuzz":
		cmd = exec.Command("exiftool", "-all=", "-overwrite_original", "-Make="+GenerateRandomString(10), "-Model="+GenerateRandomString(12), "-Software="+GenerateRandomString(15), tmpName)
	}

	if cmd != nil {
		out, err := cmd.CombinedOutput()
		if err != nil && !strings.Contains(string(out), "0 image files updated") {
			return fmt.Errorf("exiftool error: %v, output: %s", err, string(out))
		}
	}

	b, err := os.ReadFile(tmpName)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
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
