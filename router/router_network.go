package router

import (
	"github.com/apex/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/gin-gonic/gin"
	"github.com/pelican-dev/wings/environment"
	"github.com/pelican-dev/wings/environment/docker"
	"net/http"
)

type Network struct {
	Name      string `json:"name"`
	Driver    string `json:"driver"`
	NetworkID string `json:"network_id,omitempty"`
}

// Returns all networks on the wings daemon created by the panel
func getAllNetworks(c *gin.Context) {
	cli, err := environment.Docker()
	if err != nil {
	}

	// Only get networks with the 'pnw_*' network
	// Loosely ensures all networks returned are also in the panel database
	networks, err := cli.NetworkList(c, types.NetworkListOptions{
		Filters: filters.NewArgs(filters.Arg("name", "pnw_*")),
	})

	out := make([]Network, len(networks))
	for i, v := range networks {
		out[i] = Network{
			Name:      v.Name,
			Driver:    v.Driver,
			NetworkID: v.ID,
		}
	}

	c.JSON(http.StatusOK, out)

}

// Creates a new docker network on the wings daemon.
func postCreateNetwork(c *gin.Context) {
	var network Network

	if err := c.BindJSON(&network); err != nil {
		return
	}

	networkID, err := docker.CreateNetwork(network.Name, types.NetworkCreate{
		Driver: network.Driver,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "unable to create network"})
		return
	}

	// Returns created network information back to the panel
	// Mainly for the NetworkID as the panel will use that in other network requests
	c.JSON(http.StatusOK, Network{
		Name:      network.Name,
		Driver:    network.Driver,
		NetworkID: networkID,
	})
}

// Deletes a network from the node, only if no containers are present in the network
func deleteRemoveNetwork(c *gin.Context) {
	cli, err := environment.Docker()
	if err != nil {
	}

	var networkInfo Network
	if err := c.BindJSON(&networkInfo); err != nil {
		return
	}

	network, err := cli.NetworkInspect(c, networkInfo.NetworkID, types.NetworkInspectOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "network does not exist"})
		return
	}

	if len(network.Containers) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete network whilst servers are still in the network"})
	}

	err = docker.RemoveNetwork(networkInfo.NetworkID)
	if err != nil {
		log.Error("Error deleting network: " + err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "unable to delete network"})
	}

}
