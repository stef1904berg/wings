package router

import (
	"github.com/apex/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/gin-gonic/gin"
	"github.com/pelican-dev/wings/environment"
	"net/http"
)

type APIResponse struct {
	Name      string `json:"name"`
	Driver    string `json:"driver"`
	NetworkID string `json:"network_id"`
}

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

	out := make([]APIResponse, len(networks))
	for i, v := range networks {
		out[i] = APIResponse{
			Name:      v.Name,
			Driver:    v.Driver,
			NetworkID: v.ID,
		}
	}

	c.JSON(http.StatusOK, out)

}

// Creates a new docker network on the wings daemon.
func postCreateNetwork(c *gin.Context) {
	cli, err := environment.Docker()
	if err != nil { /* send error */
	}

	var network Network

	if err := c.BindJSON(&network); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "The requested resource does not exist on this instance."})
		return
	}

	newNetwork, err := cli.NetworkCreate(c.Request.Context(), network.Name, types.NetworkCreate{
		Driver: network.Driver,
	})
	if err != nil {
		log.Debug(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not create network"})
		return
	}

	// Returns created network information back to the panel
	// Mainly for the NetworkID as the panel will use that in all other network requests

	c.JSON(http.StatusOK, APIResponse{
		Name:      network.Name,
		Driver:    network.Driver,
		NetworkID: newNetwork.ID,
	})
}

// Deletes a network from the node, only if no containers are present in the network
func deleteRemoveNetwork(c *gin.Context) {
	cli, err := environment.Docker()
	if err != nil {
	}

	var networkInfo Network
	if err := c.BindJSON(&networkInfo); err != nil {
	}

	network, err := cli.NetworkInspect(c, networkInfo.NetworkID, types.NetworkInspectOptions{})
	if err != nil {
		log.Debug("Tried deleting network that doesnt exist. " + err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Network does not exist"})
		return
	}

	if len(network.Containers) != 0 {
		log.Debug("Trying to delete network but network still has containers attached to it")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete network, containers are still attached"})
	}

	err = cli.NetworkRemove(c, network.ID)
	log.Info("removed network: " + network.Name)
	if err != nil {
		log.Debug("Error deleting network: " + err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Error deleting network"})
	}

}
