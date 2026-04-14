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
	registration := a.buildRegistration(cfg)

	if err := a.startServiceWatch(cfg.ServiceName); err != nil {
		return err
	}

	if err := a.client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("service registration error: %w", err)
	}

	log.Printf("Registered service: %s", cfg.ServiceName)
	return nil
}

func (a *Agent) buildRegistration(cfg Config) *api.AgentServiceRegistration {
	registration := &api.AgentServiceRegistration{
		ID:      cfg.ServiceID,
		Name:    cfg.ServiceName,
		Tags:    cfg.Tags,
		Address: cfg.Address,
		Port:    cfg.Port,
	}

	if cfg.Check != nil {
		registration.Check = cfg.Check
	}

	return registration
}

func (a *Agent) startServiceWatch(serviceName string) error {
	plan, err := watch.Parse(map[string]any{
		"type":        "service",
		"service":     serviceName,
		"passingonly": true,
	})
	if err != nil {
		return fmt.Errorf("consul watch parse error: %w", err)
	}

	plan.HybridHandler = a.handleServiceEntries

	go func() {
		if err := plan.RunWithClientAndHclog(a.client, nil); err != nil {
			log.Println("consul watch error:", err)
		}
	}()

	return nil
}

func (a *Agent) handleServiceEntries(_ watch.BlockingParamVal, result any) {
	entries, ok := result.([]*api.ServiceEntry)
	if !ok {
		return
	}

	for _, entry := range entries {
		a.processEntry(entry)
	}
}

func (a *Agent) processEntry(entry *api.ServiceEntry) {
	if entry == nil || entry.Service == nil {
		return
	}

	id := entry.Service.ID
	if a.seenInstances[id] {
		return
	}

	a.seenInstances[id] = true

	fmt.Printf(
		"🟢 New instance joined: %s (Address=%s Port=%d)\n",
		entry.Service.ID,
		entry.Service.Address,
		entry.Service.Port,
	)
}
