package main

import (
	"archive/tar"
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"os"

	docker "github.com/fsouza/go-dockerclient"
)

func main() {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	dockerfile := "docker/Dockerfile"

	// Create a buffer
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	// Create a filereader
	dockerFileReader, err := os.Open(dockerfile)
	if err != nil {
		log.Fatal(err)
	}

	// Read the actual Dockerfile
	readDockerFile, err := ioutil.ReadAll(dockerFileReader)
	if err != nil {
		log.Fatal(err)
	}

	// Make a TAR header for the file
	tarHeader := &tar.Header{
		Name: dockerfile,
		Size: int64(len(readDockerFile)),
	}

	// Writes the header described for the TAR file
	err = tw.WriteHeader(tarHeader)
	if err != nil {
		log.Fatal(err)
	}

	// Writes the dockerfile data to the TAR file
	_, err = tw.Write(readDockerFile)
	if err != nil {
		log.Fatal(err)
	}

	dockerFileTarReader := bytes.NewReader(buf.Bytes())
	opts := docker.BuildImageOptions{
		Context:      ctx,
		Name:         "hello-world",
		Dockerfile:   dockerfile,
		InputStream:  dockerFileTarReader,
		OutputStream: bytes.NewBuffer(nil),
		Pull:         false,
	}
	if err := client.BuildImage(opts); err != nil {
		log.Fatal(err)
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"80/tcp": {{HostIP: "0.0.0.0", HostPort: "8080"}}}

	createContHostConfig := docker.HostConfig{
		PortBindings:    portBindings,
		PublishAllPorts: true,
		Privileged:      false,
	}

	exposedCadvPort := map[docker.Port]struct{}{
		"80/tcp": {}}

	createContConf := docker.Config{
		ExposedPorts: exposedCadvPort,
		Image:        "hello-world:latest",
	}

	optsContainer := docker.CreateContainerOptions{
		Context:    ctx,
		Name:       "hello-world",
		Config:     &createContConf,
		HostConfig: &createContHostConfig,
	}
	container, err := client.CreateContainer(optsContainer)
	if err != nil {
		log.Fatal(err)
	}
	if err := client.StartContainerWithContext(container.ID, nil, ctx); err != nil {
		log.Fatal(err)
	}
}
