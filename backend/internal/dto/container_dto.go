package dto

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

type CreateContainerDto struct {
	Name          string            `json:"name" binding:"required"`
	Image         string            `json:"image" binding:"required"`
	Command       []string          `json:"command,omitempty"`
	Entrypoint    []string          `json:"entrypoint,omitempty"`
	WorkingDir    string            `json:"workingDir,omitempty"`
	User          string            `json:"user,omitempty"`
	Environment   []string          `json:"environment,omitempty"`
	Ports         map[string]string `json:"ports,omitempty"`
	Volumes       []string          `json:"volumes,omitempty"`
	Networks      []string          `json:"networks,omitempty"`
	RestartPolicy string            `json:"restartPolicy,omitempty"`
	Privileged    bool              `json:"privileged,omitempty"`
	AutoRemove    bool              `json:"autoRemove,omitempty"`
	Memory        int64             `json:"memory,omitempty"`
	CPUs          float64           `json:"cpus,omitempty"`
}

type ContainerActionResult struct {
	Started []string `json:"started,omitempty"`
	Stopped []string `json:"stopped,omitempty"`
	Failed  []string `json:"failed,omitempty"`
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}

type PortDto struct {
	IP          string `json:"ip,omitempty"`
	PrivatePort int    `json:"privatePort"`
	PublicPort  int    `json:"publicPort,omitempty"`
	Type        string `json:"type"`
}

type MountDto struct {
	Type        string `json:"type"`
	Name        string `json:"name,omitempty"`
	Source      string `json:"source,omitempty"`
	Destination string `json:"destination"`
	Driver      string `json:"driver,omitempty"`
	Mode        string `json:"mode,omitempty"`
	RW          bool   `json:"rw,omitempty"`
	Propagation string `json:"propagation,omitempty"`
}

type NetworkDto struct {
	IPAMConfig          any               `json:"ipamConfig"`
	Links               []string          `json:"links"`
	Aliases             []string          `json:"aliases"`
	MacAddress          string            `json:"macAddress"`
	DriverOpts          map[string]string `json:"driverOpts"`
	GwPriority          int               `json:"gwPriority"`
	NetworkID           string            `json:"networkId"`
	EndpointID          string            `json:"endpointId"`
	Gateway             string            `json:"gateway"`
	IPAddress           string            `json:"ipAddress"`
	IPPrefixLen         int               `json:"ipPrefixLen"`
	IPv6Gateway         string            `json:"ipv6Gateway"`
	GlobalIPv6Address   string            `json:"globalIPv6Address"`
	GlobalIPv6PrefixLen int               `json:"globalIPv6PrefixLen"`
	DNSNames            []string          `json:"dnsNames"`
}

type HostConfigDto struct {
	NetworkMode string `json:"networkMode"`
}

type NetworkSettingsDto struct {
	Networks map[string]NetworkDto `json:"networks"`
}

type ContainerSummaryDto struct {
	ID              string             `json:"id"`
	Names           []string           `json:"names"`
	Image           string             `json:"image"`
	ImageID         string             `json:"imageId"`
	Command         string             `json:"command"`
	Created         int64              `json:"created"`
	Ports           []PortDto          `json:"ports"`
	Labels          map[string]string  `json:"labels"`
	State           string             `json:"state"`
	Status          string             `json:"status"`
	HostConfig      HostConfigDto      `json:"hostConfig"`
	NetworkSettings NetworkSettingsDto `json:"networkSettings"`
	Mounts          []MountDto         `json:"mounts"`
}

// Mapper from Docker's container.Summary to DTO
func NewContainerSummaryDto(c container.Summary) ContainerSummaryDto {
	// Ports
	ports := make([]PortDto, 0, len(c.Ports))
	for _, p := range c.Ports {
		ports = append(ports, PortDto{
			IP:          p.IP,
			PrivatePort: int(p.PrivatePort),
			PublicPort:  int(p.PublicPort),
			Type:        p.Type,
		})
	}

	// Mounts
	mounts := make([]MountDto, 0, len(c.Mounts))
	for _, m := range c.Mounts {
		mounts = append(mounts, MountDto{
			Type:        string(m.Type),
			Name:        m.Name,
			Source:      m.Source,
			Destination: m.Destination,
			Driver:      m.Driver,
			Mode:        m.Mode,
			RW:          m.RW,
			Propagation: string(m.Propagation),
		})
	}

	// Networks
	networks := map[string]NetworkDto{}
	if c.NetworkSettings != nil && c.NetworkSettings.Networks != nil {
		for name, n := range c.NetworkSettings.Networks {
			networks[name] = mapEndpointSettings(n)
		}
	}

	return ContainerSummaryDto{
		ID:      c.ID,
		Names:   c.Names,
		Image:   c.Image,
		ImageID: c.ImageID,
		Command: c.Command,
		Created: c.Created,
		Ports:   ports,
		Labels:  c.Labels,
		State:   c.State,
		Status:  c.Status,
		HostConfig: HostConfigDto{
			NetworkMode: c.HostConfig.NetworkMode,
		},
		NetworkSettings: NetworkSettingsDto{
			Networks: networks,
		},
		Mounts: mounts,
	}
}

func mapEndpointSettings(n *network.EndpointSettings) NetworkDto {
	if n == nil {
		return NetworkDto{}
	}
	// DriverOpts can be nil in Docker types; ensure non-nil map for JSON stability if needed
	var driverOpts map[string]string
	if n.DriverOpts != nil {
		driverOpts = n.DriverOpts
	}

	return NetworkDto{
		IPAMConfig:          n.IPAMConfig,
		Links:               n.Links,
		Aliases:             n.Aliases,
		MacAddress:          n.MacAddress,
		DriverOpts:          driverOpts,
		GwPriority:          n.GwPriority,
		NetworkID:           n.NetworkID,
		EndpointID:          n.EndpointID,
		Gateway:             n.Gateway,
		IPAddress:           n.IPAddress,
		IPPrefixLen:         n.IPPrefixLen,
		IPv6Gateway:         n.IPv6Gateway,
		GlobalIPv6Address:   n.GlobalIPv6Address,
		GlobalIPv6PrefixLen: n.GlobalIPv6PrefixLen,
		DNSNames:            n.DNSNames,
	}
}
