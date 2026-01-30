package cmd

import (
	"fmt"
)

func PrintUsage() {
	fmt.Println("Redway - Redroid Container Manager")
	fmt.Println("\nUsage: redway [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  init [image]         Initialize redroid container")
	fmt.Println("  start                Start container")
	fmt.Println("  stop                 Stop container")
	fmt.Println("  restart              Restart container")
	fmt.Println("  status               Show container status")
	fmt.Println("  shell                Enter container shell")
	fmt.Println("  adb-connect          Show ADB connection command")
	fmt.Println("  remove               Remove container and data")
	fmt.Println("  list                 List available containers")
	fmt.Println("  log                  Show container logs")
	fmt.Println("\nExamples:")
	fmt.Println("  redway init                                              # Use default redroid image")
	fmt.Println("  redway init docker://redroid/redroid:16.0.0_64only-latest  # Use specific OCI image")
	fmt.Println("  redway start")
	fmt.Println("  redway adb-connect")
	fmt.Println("  redway shell")
	fmt.Println("\nDefault image: docker://redroid/redroid:16.0.0_64only-latest")
	fmt.Println("\nNote: Images are fetched from OCI registries via skopeo and converted to LXC containers.")
}
