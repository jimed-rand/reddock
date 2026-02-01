package container

import (
	"fmt"
	"reddock/pkg/config"
)

type DockerfileGenerator struct {
	config        *config.Config
	containerName string
}

func NewDockerfileGenerator(containerName string) *DockerfileGenerator {
	cfg, _ := config.Load()
	return &DockerfileGenerator{
		config:        cfg,
		containerName: containerName,
	}
}

func (g *DockerfileGenerator) Generate() (string, error) {
	container := g.config.GetContainer(g.containerName)
	if container == nil {
		return "", fmt.Errorf("Container '%s' is not found", g.containerName)
	}

	gpuParam := "auto"
	if container.GPUMode != "" {
		gpuParam = container.GPUMode
	}

	dockerfile := fmt.Sprintf("# Dockerfile for %s\n", container.Name)
	dockerfile += fmt.Sprintf("FROM %s\n\n", container.ImageURL)

	dockerfile += "# Redroid boot arguments\n"
	dockerfile += fmt.Sprintf("CMD [\"androidboot.redroid_gpu_mode=%s\"]\n", gpuParam)

	return dockerfile, nil
}

func (g *DockerfileGenerator) Show() error {
	dockerfile, err := g.Generate()
	if err != nil {
		return err
	}

	fmt.Println(dockerfile)
	return nil
}
