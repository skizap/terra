package installer

// Installer is the interface components must use
type Installer interface {
	// Install performs the installation
	Install() error
}
