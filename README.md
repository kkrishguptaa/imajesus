# imajesus

Imajesus is a CLI tool built to optimize images for the web. It converts images to WebP format, strips metadata, and optionally resizes them.

Think of it as an image optimizer but in your terminal. You can convert PNGs, JPEGs to WebP, resize them, and strip metadata — all in one command.

## 📦 Installation

### With Golang

```sh
go install github.com/kkrishguptaa/imajesus@latest
```

### Without Golang

1. Download the release file from the [latest release](https://github.com/kkrishguptaa/imajesus/releases/latest)
1. Place it in any folder under `$PATH`

## ✌️ Usage

### File

```sh
# Convert a single image to WebP
imajesus file image.png

# Convert with a custom output name
imajesus file image.png output.webp

# Convert with custom quality (default: 60)
imajesus -q 70 file image.png

# Convert and resize to 1800px wide (preserving aspect ratio)
imajesus --resize 1800x file image.png

# Convert, resize, and delete the original
imajesus --resize 1800x -d file image.png

# Resize to a specific height
imajesus --resize x600 file image.png

# Resize to exact dimensions
imajesus --resize 1800x600 file image.png
```

### Directory

```sh
# Convert all compatible images in a directory to WebP
imajesus dir ./images

# Convert all images with custom quality
imajesus -q 80 dir ./images

# Convert, resize, and delete originals
imajesus --resize 1800x -d dir ./images
```

The `dir` command processes all `.png`, `.jpg`, and `.jpeg` files in the directory and displays a progress bar while converting.

### Flags

| Flag        | Short | Default  | Description                                           |
| ----------- | ----- | -------- | ----------------------------------------------------- |
| `--quality` | `-q`  | `60`     | Quality to pass to the WebP encoder (0-100)           |
| `--resize`  | `-r`  | disabled | Geometry for resize, e.g. `1800x`, `x600`, `1800x600` |
| `--delete`  | `-d`  | `false`  | Delete the original input image after conversion      |

### Pipeline

1. **Resize** (optional) — Uses ImageMagick to resize the image to the specified geometry, saving to a temp file (`~filename`), and strips metadata
2. **Convert** — Decodes the source image (PNG/JPEG) and encodes it to WebP at the specified quality — metadata is inherently stripped
3. **Cleanup** — Removes the temp resized file, and optionally deletes the original
