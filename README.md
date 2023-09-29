# go-manifest

`go-manifest` is a Go library inspired by [manifestival](https://github.com/manifestival/manifestival), but with a focus
on using Kubernetes Server-Side Apply. Server-Side Apply is a powerful feature introduced in Kubernetes that allows you
to declaratively update resources on the server side, which can be more efficient and reliable than traditional
imperative updates.

## Installation

To use `go-manifest` in your Go project, you can simply add it as a dependency:

```shell
go get github.com/totvs-cloud/go-manifest
```

## Usage

Here's a basic example of how to use `go-manifest`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/totvs-cloud/go-manifest"
)

func main() {
	// Create a new Manifest object
	m, err := manifest.NewManifest("path/to/your/manifest.yaml")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Apply the manifest using Server-Side Apply
	err = m.Apply()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Manifest applied successfully!")
}
```

In this example, we create a `Manifest` object by specifying the path to your Kubernetes manifest file. Then, we use
the `Apply` method to apply the manifest using Server-Side Apply.

## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue on the GitHub repository.

## Acknowledgments

This library is inspired by [manifestival](https://github.com/manifestival/manifestival) and leverages the power of
Kubernetes Server-Side Apply. We are grateful to the open-source community and the Kubernetes project for making these
tools available.
