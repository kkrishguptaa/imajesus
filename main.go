package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chai2010/webp"
	"github.com/spf13/cobra"
	"gopkg.in/gographics/imagick.v3/imagick"
)

var (
	quality    int
	resizeGeom string
	deleteOrig bool
)

var rootCmd = &cobra.Command{
	Use:   "imajesus",
	Short: "Don't fret about image performance on the web",
	Long:  "Imajesus converts images to WebP, strips metadata, and optionally resizes them.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var fileCmd = &cobra.Command{
	Use:   "file <input> [output]",
	Short: "Process a single image file",
	Long:  "Convert a single image to WebP, strip metadata and optionally resize with ImageMagick.",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		output := webpName(input)
		if len(args) == 2 {
			output = args[1]
		}
		return processFile(input, output)
	},
}

var dirCmd = &cobra.Command{
	Use:   "dir <directory>",
	Short: "Process a directory of images",
	Long:  "Convert all compatible images in a directory to WebP, strip metadata, and optionally resize with ImageMagick.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return processDir(args[0])
	},
}

// parseGeometry parses a geometry string like "1800x", "x600", or "1800x600".
// A zero value means "compute from aspect ratio".
func parseGeometry(geom string) (uint, uint, error) {
	geom = strings.TrimSpace(geom)
	parts := strings.SplitN(geom, "x", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid geometry %q: expected WxH, Wx, or xH", geom)
	}

	var w, h uint
	if parts[0] != "" {
		v, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid width in geometry %q: %w", geom, err)
		}
		w = uint(v)
	}
	if parts[1] != "" {
		v, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid height in geometry %q: %w", geom, err)
		}
		h = uint(v)
	}
	if w == 0 && h == 0 {
		return 0, 0, fmt.Errorf("geometry %q must specify at least width or height", geom)
	}
	return w, h, nil
}

// resolveSize computes final dimensions, preserving aspect ratio when one is zero.
func resolveSize(origW, origH, targetW, targetH uint) (uint, uint) {
	if targetW == 0 {
		targetW = uint(float64(origW) * float64(targetH) / float64(origH))
	}
	if targetH == 0 {
		targetH = uint(float64(origH) * float64(targetW) / float64(origW))
	}
	return targetW, targetH
}

// ── Core pipeline ───────────────────────────────────────────────────────────
// 1. Resize with ImageMagick via imagick (optional) → temp file ~input
// 2. Decode image → encode to WebP with chai2010/webp
// 3. Delete original (optional)

func processFile(input, output string) error {
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", input)
	}

	source := input

	// Step 1 – resize with ImageMagick if requested
	if resizeGeom != "" {
		targetW, targetH, err := parseGeometry(resizeGeom)
		if err != nil {
			return err
		}

		dir := filepath.Dir(input)
		base := filepath.Base(input)
		resized := filepath.Join(dir, "~"+base)

		imagick.Initialize()
		defer imagick.Terminate()

		mw := imagick.NewMagickWand()
		defer mw.Destroy()

		if err := mw.ReadImage(input); err != nil {
			return fmt.Errorf("imagick read failed: %w", err)
		}

		w, h := resolveSize(mw.GetImageWidth(), mw.GetImageHeight(), targetW, targetH)
		if err := mw.ResizeImage(w, h, imagick.FILTER_LANCZOS); err != nil {
			return fmt.Errorf("imagick resize failed: %w", err)
		}

		if err := mw.StripImage(); err != nil {
			return fmt.Errorf("imagick strip failed: %w", err)
		}

		if err := mw.WriteImage(resized); err != nil {
			return fmt.Errorf("imagick write failed: %w", err)
		}

		source = resized
		defer os.Remove(resized)
	}

	// Step 2 – decode source image and encode to WebP
	srcFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	img, _, err := image.Decode(srcFile)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Quality: float32(quality)}); err != nil {
		return fmt.Errorf("webp encode failed: %w", err)
	}

	if err := os.WriteFile(output, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Step 3 – delete original if requested
	if deleteOrig {
		if err := os.Remove(input); err != nil {
			return fmt.Errorf("failed to delete original: %w", err)
		}
	}

	return nil
}

func processDir(dir string) error {
	jobs, err := LoopDirectory(dir)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		fmt.Println("No compatible images found in", dir)
		return nil
	}
	return RunProgressUI(jobs)
}

func main() {
	rootCmd.PersistentFlags().IntVarP(&quality, "quality", "q", 60, "Quality to pass to webp encoder (0-100)")
	rootCmd.PersistentFlags().StringVarP(&resizeGeom, "resize", "r", "", "Geometry for resize (e.g. 1800x, x600, 1800x600)")
	rootCmd.PersistentFlags().BoolVarP(&deleteOrig, "delete", "d", false, "Delete the original input image after conversion")

	rootCmd.AddCommand(fileCmd)
	rootCmd.AddCommand(dirCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
