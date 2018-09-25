module github.com/stellarproject/terra

require (
	github.com/codegangsta/cli v1.20.0
	github.com/containerd/containerd v1.1.3
	github.com/containerd/continuity v0.0.0-20180921161001-7f53d412b9eb // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/mitchellh/go-homedir v1.0.0
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/pkg/errors v0.8.0
	github.com/pkg/sftp v1.8.3
	github.com/sirupsen/logrus v1.0.6
	golang.org/x/crypto v0.0.0-20180910181607-0e37d006457b
	golang.org/x/net v0.0.0-20180921000356-2f5d2388922f // indirect
	golang.org/x/sys v0.0.0-20180920110915-d641721ec2de // indirect
	google.golang.org/grpc v1.15.0 // indirect
)

replace github.com/containerd/containerd => github.com/containerd/containerd v1.2.0-rc.0.0.20180921182233-87d1118a0f35
