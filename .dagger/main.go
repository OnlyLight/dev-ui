package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

// DevUIPipeline defines the Dagger pipeline for the dev-ui project
type DevUIPipeline struct {
	client *dagger.Client // Shared Dagger client
}

// NewDevUIPipeline initializes the pipeline with a Dagger client
func NewDevUIPipeline(ctx context.Context) (*DevUIPipeline, error) {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return nil, err
	}
	return &DevUIPipeline{client: client}, nil
}

// LoadSource loads the source code from the local directory
func (p *DevUIPipeline) LoadSource(ctx context.Context) *dagger.Directory {
	return p.client.Host().Directory("./website", dagger.HostDirectoryOpts{
		Exclude: []string{".git", "node_modules"}, // Ignore unnecessary files
	})
}

// BuildImage builds the Docker image using the existing Dockerfile
func (p *DevUIPipeline) BuildImage(ctx context.Context, source *dagger.Directory) *dagger.Container {
	return p.client.Container().Build(source, dagger.ContainerBuildOpts{
		Dockerfile: "Dockerfile",
	})
}

// Publish pushes the Docker image to the registry with a given tag
func (p *DevUIPipeline) Publish(ctx context.Context, container *dagger.Container, dockerURL, dockerUsername, dockerPassword, tag string) (string, error) {
	return container.
		WithRegistryAuth(dockerURL, dockerUsername, p.client.SetSecret("docker-password", dockerPassword)).
		Publish(ctx, fmt.Sprintf("%s/crawler-website:%s", dockerUsername, tag))
}

func (p *DevUIPipeline) Getenv(ctx context.Context) (string, string, string, string) {
	dockerUsername := os.Getenv("DOCKER_USERNAME")
	dockerPassword := os.Getenv("DOCKER_PASSWORD")
	dockerURL := os.Getenv("DOCKER_URL")
	imageTag := os.Getenv("IMAGE_TAG")
	if dockerURL == "" {
		dockerURL = "docker.io" // Default to Docker Hub
	}
	if dockerUsername == "" || dockerPassword == "" {
		panic("DOCKER_USERNAME and DOCKER_PASSWORD must be set")
	}

	return dockerUsername, dockerPassword, dockerURL, imageTag
}

// Pipeline executes the full Dagger pipeline for dev-ui
func (p *DevUIPipeline) Pipeline(ctx context.Context) (string, string, error) {
	// Step 1: Load source code
	source := p.LoadSource(ctx)

	// Step 2: Get env variable
	dockerUsername, dockerPassword, dockerURL, imageTag := p.Getenv(ctx)

	// Step 3: Build the Docker image
	container := p.BuildImage(ctx, source)

	// Step 4: Publish the Docker image with commit tag
	imageRef, err := p.Publish(ctx, container, dockerURL, dockerUsername, dockerPassword, imageTag)
	if err != nil {
		return "", "", err
	}

	// Step 5: Publish the Docker image with 'latest' tag
	latestRef, err := p.Publish(ctx, container, dockerURL, dockerUsername, dockerPassword, "latest")
	if err != nil {
		return "", "", err
	}

	return imageRef, latestRef, nil
}

func main() {
	// Initialize a context
	ctx := context.Background()

	// Initialize the pipeline
	pipeline, err := NewDevUIPipeline(ctx)
	if err != nil {
		panic(err)
	}
	defer pipeline.client.Close()

	// Run the pipeline
	imageRef, latestRef, err := pipeline.Pipeline(ctx)
	if err != nil {
		panic(err)
	}

	// Print the results
	fmt.Printf("Pushed image: %s\n", imageRef)
	fmt.Printf("Pushed image: %s\n", latestRef)
}
