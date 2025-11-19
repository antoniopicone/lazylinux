package image

import (
	"fmt"
	"io"
	"lazylinux-vm/internal/config"
	"lazylinux-vm/internal/utils"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func GetImageURL(imageName, arch string) (string, error) {
	switch imageName {
	case "debian13":
		if arch == "amd64" || arch == "x86_64" {
			return "https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-amd64.qcow2", nil
		}
		return "https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-arm64.qcow2", nil
	case "debian12":
		if arch == "amd64" || arch == "x86_64" {
			return "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-generic-amd64.qcow2", nil
		}
		return "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-arm64.qcow2", nil
	default:
		return "", fmt.Errorf("unknown image: %s", imageName)
	}
}

func DownloadImage(imageName, arch string) (string, error) {
	url, err := GetImageURL(imageName, arch)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s-%s.qcow2", imageName, arch)
	dest := filepath.Join(config.ImagesDir, filename)

	if _, err := os.Stat(dest); err == nil {
		return dest, nil
	}

	if err := os.MkdirAll(config.ImagesDir, 0755); err != nil {
		return "", err
	}

	fmt.Printf("%s[i]%s Downloading %s for %s architecture...\n", utils.Blue, utils.NC, imageName, arch)

	tempFile := dest + ".tmp"
	out, err := os.Create(tempFile)
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	// Convert if needed (simple check based on URL extension, though logic here assumes qcow2 mostly)
	// The original script handles .img conversion. For now, we assume qcow2 download.

	if err := os.Rename(tempFile, dest); err != nil {
		return "", err
	}

	// Verify with qemu-img info
	cmd := exec.Command("qemu-img", "info", dest)
	if err := cmd.Run(); err != nil {
		os.Remove(dest)
		return "", fmt.Errorf("downloaded image is not valid")
	}

	return dest, nil
}
