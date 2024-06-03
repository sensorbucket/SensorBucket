package userworkers

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
)

type DockerController struct {
	docker client.APIClient
	store  Store

	server *http.Server

	workersEndpoint string
	workerImage     string
	workerNetID     string
	amqpHost        string
	amqpExchange    string
	endpointDevices string
}

func CreateDockerController(store Store) (*DockerController, error) {
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	// Find network id
	netID := env.Could("CTRL_DOCKER_WORKER_NET", "")
	if netID == "" {
		// try to get it ourselves
		nets, err := docker.NetworkList(context.Background(), types.NetworkListOptions{})
		if err != nil {
			return nil, fmt.Errorf("not network id given and an error occured while trying to get a network: %w", err)
		}
		for _, net := range nets {
			if strings.Contains(strings.ToLower(net.Name), "sensorbucket") {
				netID = net.ID
				break
			}
		}
	}
	ctrl := &DockerController{
		docker: docker,
		store:  store,

		workersEndpoint: env.Could("CTLR_DOCKER_WORKERS_EP", "http://userworkers:3001"),
		workerImage:     env.Could("CTLR_DOCKER_WORKER_IMAGE", "sensorbucket/docker-worker:latest"),
		workerNetID:     netID,
		amqpHost:        env.Could("CTRL_DOCKER_AMQP_HOST", "amqp://guest:guest@mq:5672"),
		amqpExchange:    env.Could("CTRL_DOCKER_AMQP_XCHG", "pipeline.messages"),
		endpointDevices: env.Could("CTRL_DOCKER_ENDPOINT_DEVICES", "http://caddy/api/devices"),
	}

	// Setup an unprotected route for the dockerworkers
	router := chi.NewRouter()
	router.Get("/{workerID}/source", func(w http.ResponseWriter, r *http.Request) {
		workerID := chi.URLParam(r, "workerID")
		worker, err := store.GetWorkerByID(uuid.MustParse(workerID), WorkerFilters{})
		if err != nil {
			log.Printf("In docker controller worker http route: failed to get worker by ID: %s\n", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		src := base64.StdEncoding.EncodeToString(worker.ZipSource)
		web.HTTPResponse(w, http.StatusOK, web.APIResponseAny{
			Data: src,
		})
	})
	ctrl.server = &http.Server{
		Addr:    "0.0.0.0:3001",
		Handler: router,
	}
	go func() {
		if err := ctrl.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Docker Controller server errored: %s\n", err)
		}
	}()

	return ctrl, nil
}

func (ctrl *DockerController) Shutdown(ctx context.Context) error {
	return ctrl.server.Shutdown(ctx)
}

func (ctrl *DockerController) Reconcile(ctx context.Context) error {
	log.Println("Starting docker reconciliation...")
	defer log.Println("Reconciliation finished")
	filter := filters.NewArgs(
		filters.Arg("label", "controlled-by=sensorbucket"),
	)
	containers, err := ctrl.docker.ContainerList(ctx, types.ContainerListOptions{
		Filters: filter,
	})
	if err != nil {
		return fmt.Errorf("cannot list containers: %w", err)
	}

	// Remove wandering containers
	containerWorkerIDs := []uuid.UUID{}
	containerWorkerIDMap := map[uuid.UUID]types.Container{}
	for _, c := range containers {
		workerIDStr := c.Labels["worker-id"]
		workerID, err := uuid.Parse(workerIDStr)
		if err != nil {
			log.Printf("Container (%s) has invalid worker-id: %s\n", c.ID, workerIDStr)
			continue
		}
		containerWorkerIDMap[workerID] = c
		containerWorkerIDs = append(containerWorkerIDs, workerID)
	}
	existingIDs, err := ctrl.store.WorkersExists(containerWorkerIDs, WorkerFilters{State: StateEnabled})
	if err != nil {
		return fmt.Errorf("error fetching which workers exist from store: %w", err)
	}
	wandering, _ := lo.Difference(containerWorkerIDs, existingIDs)
	log.Printf("Removing %d wandering containers\n", len(wandering))
	for _, id := range wandering {
		c := containerWorkerIDMap[id]
		err := ctrl.docker.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{Force: true})
		if err != nil {
			log.Printf("Error removing container: %s: %v\n", c.ID, err)
			continue
		}
	}

	// Iterate over workers in Database
	var cursor string
	for {
		pages, err := ctrl.store.ListUserWorkers(WorkerFilters{State: StateEnabled}, pagination.Request{Limit: 10, Cursor: cursor})
		if err != nil {
			return fmt.Errorf("error listing user workers from database: %w", err)
		}

		toCreate := []UserWorker{}
		toUpdate := []UserWorker{}
		for _, worker := range pages.Data {
			if worker.Language != LanguagePython {
				continue
			}
			filter := filters.NewArgs(
				filters.Arg("label", "controlled-by=sensorbucket"),
				filters.Arg("label", "worker-id="+worker.ID.String()),
			)
			candidates, err := ctrl.docker.ContainerList(ctx, types.ContainerListOptions{
				All:     true,
				Filters: filter,
			})
			if err != nil {
				log.Printf("Error getting containers for worker-id: %s: %v\n", worker.ID.String(), err)
				continue
			}
			// Must be created
			if len(candidates) == 0 {
				log.Printf("WHY: creating %s because it has no instances\n", worker.ID.String())
				toCreate = append(toCreate, worker)
				continue
			} else if len(candidates) > 1 {
				// Updating is deleting then creating, so this will remove all duplicates and then create a single
				log.Printf("WHY: updating %s because it has too many instances\n", worker.ID.String())
				toUpdate = append(toUpdate, worker)
				continue
			}
			container := candidates[0]
			revision, err := strconv.ParseInt(container.Labels["worker-revision"], 10, 64)
			if err != nil {
				log.Printf("Error converting worker revision to int for container: %s\n", container.ID)
				continue
			}
			if revision < int64(worker.Revision) {
				log.Printf("WHY: updating %s because it's revision is lower (%d < %d)\n", worker.ID.String(), revision, worker.Revision)
				toUpdate = append(toUpdate, worker)
				continue
			}
			if !lo.Contains([]string{"running", "starting"}, container.State) {
				log.Printf("WHY: updating %s because it's state dictates so: %s\n", worker.ID.String(), container.State)
				toUpdate = append(toUpdate, worker)
			}
		}

		// Commit changes
		log.Printf("Creating %d workers...\n", len(toCreate))
		for _, worker := range toCreate {
			if err := ctrl.createContainerForWorker(ctx, worker); err != nil {
				log.Printf("Error creating container for worker: %v\n", err)
			}
		}
		log.Printf("Updating %d workers...\n", len(toUpdate))
		for _, worker := range toUpdate {
			if err := ctrl.removeContainer(ctx, worker); err != nil {
				log.Printf("Error removing container for worker: %v\n", err)
				continue
			}
			if err := ctrl.createContainerForWorker(ctx, worker); err != nil {
				log.Printf("Error creating container for worker: %v\n", err)
			}
		}

		// Continue to next page if there is one
		if pages.Cursor == "" {
			break
		}
		cursor = pagination.CursorFromURL(pages.Cursor)
	}

	return nil
}

func (ctrl *DockerController) createContainerForWorker(ctx context.Context, worker UserWorker) error {
	cfg := &container.Config{
		Labels: map[string]string{
			"controlled-by":   "sensorbucket",
			"worker-id":       worker.ID.String(),
			"worker-revision": strconv.FormatInt(int64(worker.Revision), 10),
		},
		Image: ctrl.workerImage,
		Env: []string{
			fmt.Sprintf("WORKER_ID=%s", worker.ID.String()),
			fmt.Sprintf("AMQP_HOST=%s", ctrl.amqpHost),
			fmt.Sprintf("AMQP_XCHG=%s", ctrl.amqpExchange),
			fmt.Sprintf("CODE_URL=%s/%s/source", ctrl.workersEndpoint, worker.ID.String()),
			fmt.Sprintf("CODE_ENTRYPOINT=%s", worker.Entrypoint),
			fmt.Sprintf("ENDPOINT_DEVICES=%s", ctrl.endpointDevices),
		},
		Tty: true,
	}
	hostCfg := &container.HostConfig{
		AutoRemove: false,
	}
	netCfg := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"sensorbucket": {
				NetworkID: ctrl.workerNetID,
			},
		},
	}
	res, err := ctrl.docker.ContainerCreate(ctx, cfg, hostCfg, netCfg, nil, fmt.Sprintf("worker-%s", worker.ID.String()))
	if err != nil {
		return fmt.Errorf("error creating new container for worker %s: %w", worker.ID.String(), err)
	}
	if len(res.Warnings) > 0 {
		log.Printf("Warnings creating worker (%s): \n%s\n", worker.ID.String(), strings.Join(res.Warnings, "\n"))
	}
	err = ctrl.docker.ContainerStart(ctx, res.ID, types.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("error starting container: %s: %w", worker.ID.String(), err)
	}
	return nil
}

func (ctrl *DockerController) removeContainer(ctx context.Context, worker UserWorker) error {
	filter := filters.NewArgs(
		filters.Arg("label", "controlled-by=sensorbucket"),
		filters.Arg("label", fmt.Sprintf("worker-id=%s", worker.ID.String())),
	)
	containers, err := ctrl.docker.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return fmt.Errorf("error listing docker containers: %w", err)
	}
	for _, container := range containers {
		if err := ctrl.docker.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
			log.Printf("Error removing container: %s\n", container.ID)
		}
	}
	return nil
}
