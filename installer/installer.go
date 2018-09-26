package installer

import (
	"io/ioutil"
	"os"
	"os/exec"
)

// Installer is the interface components must use
type Installer interface {
	// Install performs the installation
	Install() ([]byte, error)
}

// AssemblyInstaller holds the information necessary for installation
type AssemblyInstaller struct {
	Image string
}

// Install performs the assembly installation
func (a *AssemblyInstaller) Install() error {
	tmpdir, err := ioutil.TempDir("", "terra-assembly-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)

	if err := FetchImage(a.Image, tmpdir); err != nil {
		return err
	}

	// exec 'install' from package
	cmd := exec.Command("./install")
	cmd.Dir = tmpdir
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
