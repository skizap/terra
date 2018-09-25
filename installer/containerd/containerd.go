package containerd

import (
	"github.com/stellarproject/terra/installer"
)

type ContainerdInstaller struct {
	Image    string
	DestPath string
}

func (c *ContainerdInstaller) Install() error {
	if err := installer.FetchImage(c.Image, c.DestPath); err != nil {
		return err
	}

	return nil
}
