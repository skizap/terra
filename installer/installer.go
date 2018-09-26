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
func (a *AssemblyInstaller) Install() ([]byte, error) {
	tmpdir, err := ioutil.TempDir("", "terra-assembly-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpdir)

	if err := FetchImage(a.Image, tmpdir); err != nil {
		return nil, err
	}

	// exec 'install' from package
	cmd := exec.Command("./install")
	cmd.Dir = tmpdir
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return out, nil
}
