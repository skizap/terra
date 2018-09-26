# Terra

![terra](terra.png)

Terra handles node management for the Stellar Project.

# Getting Started
To build, run `make`.  This will create a `bin/terra`.

To initialize a node, run `terra bootstrap init <ssh-host>`.  This will install the base components for node management.

# Kernel
Terra supports building a kernel for a strong foundation across all nodes.

## Building

Use a provided `config` or copy your own into the current directory.

```bash
$> make kernel
```

This will place a `linux-headers.deb` and `linux-image.deb` in the current directory.

[Photo](https://www.pexels.com/photo/astronomy-atmosphere-earth-exploration-220201/)
