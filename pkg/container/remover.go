package container

import (
	"fmt"
	"os"
	"reddock/pkg/config"
)

type Remover struct {
	config        *config.Config
	containerName string
	runtime       Runtime
}

func NewRemover(containerName string) *Remover {
	cfg, _ := config.Load()
	return &Remover{
		config:        cfg,
		containerName: containerName,
		runtime:       NewRuntime(),
	}
}

func (r *Remover) Remove(removeImage bool) error {
	if err := CheckRoot(); err != nil {
		return err
	}

	container := r.config.GetContainer(r.containerName)
	if container == nil {
		return fmt.Errorf("Container '%s' not found", r.containerName)
	}

	fmt.Printf("[*] Stopping and removing container '%s'...\n", container.Name)

	// Stop if running
	if r.runtime.IsRunning(container.Name) {
		r.runtime.Stop(container.Name)
	}

	// Remove container
	if err := r.runtime.Remove(container.Name, true); err != nil {
		fmt.Printf("Warning: Failed to remove container: %v\n", err)
	}

	// Remove data directory
	dataPath := container.GetDataPath()
	fmt.Printf("[*] Removing data directory: %s\n", dataPath)
	if err := os.RemoveAll(dataPath); err != nil {
		fmt.Printf("Warning: Could not remove data directory: %v\n", err)
	}

	// Remove image if requested
	if removeImage {
		fmt.Printf("[*] Removing image: %s\n", container.ImageURL)
		if err := r.runtime.RemoveImage(container.ImageURL); err != nil {
			fmt.Printf("Warning: Could not remove image: %v\n", err)
		}
	}

	// Update config
	r.config.RemoveContainer(container.Name)
	if err := config.Save(r.config); err != nil {
		return fmt.Errorf("Failed to save config: %v", err)
	}

	fmt.Printf("\nContainer '%s' and its data have been removed successfully.\n", container.Name)
	if removeImage {
		fmt.Println("Image has also been removed.")
	}
	return nil
}
