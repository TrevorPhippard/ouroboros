package consul

import (
	"fmt"
	"log"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

type Config struct {
	ServiceID   string
	ServiceName string
	Address     string
	Port        int
	Tags        []string

	// NEW: allow passing full check config
	Check *api.AgentServiceCheck
}

type Agent struct {
	client        *api.Client
	seenInstances map[string]bool
}

func NewAgent(consulConfig *api.Config) *Agent {
	client, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatal("consul client error:", err)
	}

	return &Agent{
		client:        client,
		seenInstances: make(map[string]bool),
	}
}

/* -------------------------------------------------------------------------- */
/*                             SERVICE REGISTRATION                           */
/* -------------------------------------------------------------------------- */

func (a *Agent) RegisterService(cfg Config) error {
	// Service registration
	registration := &api.AgentServiceRegistration{
		ID:      cfg.ServiceID,
		Name:    cfg.ServiceName,
		Tags:    cfg.Tags,
		Address: cfg.Address,
		Port:    cfg.Port,
	}

	// Attach health check if provided
	if cfg.Check != nil {
		registration.Check = cfg.Check
	}

	/* ---------------------------------------------------------------------- */
	/*                         WATCH FOR NEW SERVICE INSTANCES                */
	/* ---------------------------------------------------------------------- */

	query := map[string]any{
		"type":        "service",
		"service":     cfg.ServiceName,
		"passingonly": true,
	}

	plan, err := watch.Parse(query)
	if err != nil {
		return fmt.Errorf("consul watch parse error: %w", err)
	}

	plan.HybridHandler = func(_ watch.BlockingParamVal, result any) {
		entries, ok := result.([]*api.ServiceEntry)
		if !ok {
			return
		}

		for _, entry := range entries {
			if entry == nil || entry.Service == nil {
				continue
			}

			id := entry.Service.ID
			if !a.seenInstances[id] {
				a.seenInstances[id] = true
				fmt.Printf(
					"🟢 New instance joined: %s (Address=%s Port=%d)\n",
					entry.Service.ID,
					entry.Service.Address,
					entry.Service.Port,
				)
			}
		}
	}

	// Run watch using SAME consul config (important fix)
	go func() {
		if err := plan.RunWithClientAndHclog(a.client, nil); err != nil {
			log.Println("consul watch error:", err)
		}
	}()

	// Register service
	if err := a.client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("service registration error: %w", err)
	}

	log.Printf("Registered service: %s", cfg.ServiceName)

	return nil
}