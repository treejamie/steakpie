package webhook

// RegistryPackageEvent represents a GitHub registry_package webhook payload.
type RegistryPackageEvent struct {
	Action          string          `json:"action"`
	RegistryPackage RegistryPackage `json:"registry_package"`
	Repository      Repository      `json:"repository"`
	Sender          Sender          `json:"sender"`
}

// RegistryPackage contains package information.
type RegistryPackage struct {
	Name           string         `json:"name"`
	Ecosystem      string         `json:"ecosystem"`
	PackageVersion PackageVersion `json:"package_version"`
}

// PackageVersion contains version-specific information.
type PackageVersion struct {
	ID                int64             `json:"id"`
	Version           string            `json:"version"`
	PackageURL        string            `json:"package_url"`
	ContainerMetadata ContainerMetadata `json:"container_metadata"`
}

// ContainerMetadata contains container-specific metadata.
type ContainerMetadata struct {
	Tag Tag `json:"tag"`
}

// Tag contains tag information for container images.
type Tag struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
}

// Repository contains repository information.
type Repository struct {
	FullName string `json:"full_name"`
}

// Sender contains information about who triggered the event.
type Sender struct {
	Login string `json:"login"`
}
